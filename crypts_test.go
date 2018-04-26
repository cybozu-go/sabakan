package sabakan

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
)

func testCryptsGet(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	handler := mux.NewRouter()
	InitCrypts(handler.PathPrefix("/api/v1/").Subrouter(), &EtcdClient{etcd, prefix})

	serial := "1"
	diskPath := "exists-path"
	key := "aaa"
	_, err = etcd.Put(context.Background(), path.Join(prefix, EtcdKeyCrypts, serial, diskPath), key)
	if err != nil {
		t.Fatal(err)
	}

	testData := []struct {
		path   string
		status int
		key    string
	}{
		{diskPath, 200, key},
		{"not-exist", 404, ""},
	}

	for _, td := range testData {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", path.Join("/api/v1/crypts", serial, td.path), nil)
		handler.ServeHTTP(w, r)
		resp := w.Result()

		if resp.StatusCode != td.status {
			t.Error("wrong status code, expects:", td.status, ", actual:", resp.StatusCode)
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		respKey := string(data)
		if len(td.key) > 0 && td.key != respKey {
			t.Error("wrong key, expects:", td.key, ", actual:", respKey)
		}
	}
}

func testCryptsPut(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	handler := mux.NewRouter()
	InitCrypts(handler.PathPrefix("/api/v1/").Subrouter(), &EtcdClient{etcd, prefix})

	serial := "1"
	diskPath := "put-path"

	testData := []struct {
		path   string
		status int
		key    string
	}{
		{diskPath, 201, "aaa"},
		{diskPath, 409, "bbb"},
		{"another-path", 201, string([]byte{0, 1, 2, 100, 50, 200})},
	}

	for _, td := range testData {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", path.Join("/api/v1/crypts", serial, td.path),
			strings.NewReader(td.key))
		handler.ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != td.status {
			t.Error("wrong status code, expects:", td.status, ", actual:", resp.StatusCode)
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 201 {
			continue
		}

		var respJSON struct {
			Status int    `json:"status"`
			Path   string `json:"path"`
		}
		err = json.Unmarshal(data, &respJSON)
		if err != nil {
			t.Error("invalid JSON:", string(data))
			continue
		}
		if respJSON.Status != 201 {
			t.Error("invalid status in JSON:", respJSON.Status)
		}
		if respJSON.Path != td.path {
			t.Error("invalid path in JSON:", respJSON.Path)
		}

		etcdResp, err := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyCrypts, serial, td.path))
		if len(etcdResp.Kvs) != 1 {
			t.Fatal("key is not stored in etcd, path:", td.path)
		}
		storedKey := string(etcdResp.Kvs[0].Value)
		if td.key != storedKey {
			t.Error("stored key is wrong, expect:", td.key, ", actual:", storedKey)
		}
	}
}

func testCryptsDelete(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	handler := mux.NewRouter()
	InitCrypts(handler.PathPrefix("/api/v1/").Subrouter(), &EtcdClient{etcd, prefix})

	expected := make(map[string]struct{})
	serial := "abc"
	key := "aaa"
	for i := 0; i < 5; i++ {
		diskPath := fmt.Sprintf("path%d", i)
		expected[diskPath] = struct{}{}
		target := path.Join(prefix, EtcdKeyCrypts, serial, diskPath)
		etcd.Put(context.Background(), target, key)
	}

	// dummy data to test bug in delete logic.
	serial2 := "abcd"
	target2 := path.Join(prefix, EtcdKeyCrypts, serial2, "path1")
	etcd.Put(context.Background(), target2, key)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", path.Join("/api/v1/crypts", serial), nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatal("expected: 200, actual:", resp.StatusCode)
	}

	var deletedPaths []string
	err = json.NewDecoder(resp.Body).Decode(&deletedPaths)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	actual := make(map[string]struct{})
	for _, p := range deletedPaths {
		actual[p] = struct{}{}
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatal("unexpected response:", deletedPaths)
	}

	testTarget := path.Join(prefix, EtcdKeyCrypts, serial) + "/"
	etcdResp, err := etcd.Get(context.Background(), testTarget, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	if etcdResp.Count != 0 {
		t.Fatal("expected: 0, actual:", etcdResp.Count)
	}

	testTarget2 := path.Join(prefix, EtcdKeyCrypts, serial2) + "/"
	etcdResp, err = etcd.Get(context.Background(), testTarget2, clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	if etcdResp.Count != 1 {
		t.Fatal("expected: 1, actual:", etcdResp.Count)
	}
}

func TestCrypts(t *testing.T) {
	t.Run("Get", testCryptsGet)
	t.Run("Put", testCryptsPut)
	t.Run("Delete", testCryptsDelete)
}
