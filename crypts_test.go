package sabakan

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"path"
	"reflect"
	"strconv"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
)

func TestValidatePostParams(t *testing.T) {
	valid := sabakanCrypt{"disk-a", "foo"}
	if err := validatePostParams(valid); err != nil {
		t.Fatal("validator should return nil when the args are valid.")
	}
	invalid1 := sabakanCrypt{"", "foo"}
	if err := validatePostParams(invalid1); err == nil {
		t.Fatal("validator should return error when the args are valid.")
	}
	invalid2 := sabakanCrypt{"disk-a", ""}
	if err := validatePostParams(invalid2); err == nil {
		t.Fatal("validator should return error when the args are valid.")
	}
}

func TestHandleGetCrypts(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	etcdClient := EtcdClient{etcd, prefix}
	serial := "1"
	diskPath := "exists-path"
	key := "aaa"
	inputCrypt := sabakanCrypt{Path: diskPath, Key: key}
	val, err := json.Marshal(inputCrypt)
	if err != nil {
		t.Fatal(err)
	}
	_, err = etcd.Put(context.Background(), path.Join(prefix, "crypts", serial, diskPath), string(val))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", path.Join("/api/v1/crypts", serial, diskPath), nil)
	r = mux.SetURLVars(r, map[string]string{"serial": serial, "path": diskPath})

	etcdClient.handleGetCrypts(w, r)

	resp := w.Result()
	var outputCrypt sabakanCrypt
	err = json.NewDecoder(resp.Body).Decode(&outputCrypt)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("expected: 200, actual:", resp.StatusCode)
	}
	if outputCrypt != inputCrypt {
		t.Fatal("invalid response body, ", outputCrypt)
	}
}

func TestHandleGetCryptsNotFound(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	etcdClient := EtcdClient{etcd, prefix}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/crypts", nil)
	r = mux.SetURLVars(r, map[string]string{"serial": "1", "path": "non-exists-key"})

	etcdClient.handleGetCrypts(w, r)

	resp := w.Result()
	if resp.StatusCode != 404 {
		t.Fatal("expected: 404, actual:", resp.StatusCode)
	}
}

func TestHandlePostCrypts(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	etcdClient := EtcdClient{etcd, prefix}
	serial := "1"
	diskPath := "put-path"
	key := "aaa"
	crypt := sabakanCrypt{Path: diskPath, Key: key}
	val, _ := json.Marshal(crypt)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", path.Join("/api/v1/crypts", serial), bytes.NewBuffer(val))
	r = mux.SetURLVars(r, map[string]string{"serial": serial})

	etcdClient.handlePostCrypts(w, r)

	resp := w.Result()
	var respondedCrypt sabakanCrypt
	var savedCrypt sabakanCrypt
	json.NewDecoder(resp.Body).Decode(&respondedCrypt)
	etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyCrypts, serial, diskPath))
	err = json.Unmarshal(etcdResp.Kvs[0].Value, &savedCrypt)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 201 {
		t.Fatal("expected: 201, actual:", resp.StatusCode)
	}
	if respondedCrypt != crypt {
		t.Fatal("invalid response body, ", respondedCrypt)
	}
	if savedCrypt != crypt {
		t.Fatal("saved entity is invalid, ", savedCrypt)
	}
}

func TestHandlePostCryptsInvalidBody(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	etcdClient := EtcdClient{etcd, prefix}
	w := httptest.NewRecorder()
	invalidBody, err := json.Marshal(&struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{"Taro", 1})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/api/v1/crypts/1", bytes.NewBuffer(invalidBody))
	r = mux.SetURLVars(r, map[string]string{"serial": "1"})

	etcdClient.handlePostCrypts(w, r)

	resp := w.Result()
	if w.Result().StatusCode != 400 {
		t.Fatal("expected: 400, actual:", resp.StatusCode)
	}
}

func TestHandleDeleteCrypts(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	etcdClient := EtcdClient{etcd, prefix}
	expected := deleteResponse{}
	serial := "1"
	key := "aaa"
	for i := 0; i < 5; i++ {
		diskPath := "path" + strconv.Itoa(i)
		crypt := sabakanCrypt{Path: diskPath, Key: key}
		val, err := json.Marshal(crypt)
		if err != nil {
			t.Fatal(err)
		}
		target := path.Join(prefix, EtcdKeyCrypts, serial, diskPath)
		etcd.Put(context.Background(), target, string(val))
		expected = append(expected, deletePath{target})
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", path.Join("/api/v1/crypts", serial), nil)
	r = mux.SetURLVars(r, map[string]string{"serial": serial})

	etcdClient.handleDeleteCrypts(w, r)

	resp := w.Result()
	var dresp deleteResponse
	err = json.NewDecoder(resp.Body).Decode(&dresp)
	if err != nil {
		t.Fatal(err)
	}
	etcdResp, err := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyCrypts, serial), clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("expected: 200, actual:", resp.StatusCode)
	}
	if etcdResp.Count != 0 {
		t.Fatal("expected: 0, actual:", etcdResp.Count)
	}
	if !reflect.DeepEqual(dresp, expected) {
		t.Fatal("unexpected response:", dresp)
	}
}

func TestHandleDeleteCryptsNotFound(t *testing.T) {
	etcd, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	etcdClient := EtcdClient{etcd, prefix}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/v1/crypts", nil)
	r = mux.SetURLVars(r, map[string]string{"serial": "non-exists-serial"})

	etcdClient.handleDeleteCrypts(w, r)

	resp := w.Result()
	if resp.StatusCode != 404 {
		t.Fatal("expected: 404, actual:", resp.StatusCode)
	}
}
