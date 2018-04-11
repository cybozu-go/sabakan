package sabakan

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"reflect"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
)

var (
	flagEtcdServers = flag.String("etcd-servers", "http://localhost:2379", "URLs of the backend etcd")
	flagEtcdPrefix  = flag.String("etcd-prefix", "/sabakan-test", "etcd prefix")
)

func TestMain(m *testing.M) {
	err := setupEtcd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func setupEtcd() error {
	etcd, err := newEtcdClient()
	if err != nil {
		return err
	}
	defer etcd.Close()

	_, err = etcd.Delete(context.Background(), *flagEtcdPrefix, clientv3.WithPrefix())
	return err
}

func newEtcdClient() (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(*flagEtcdServers, ","),
		DialTimeout: 2 * time.Second,
	})
}

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
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := *flagEtcdPrefix + "/TestHandleGetCrypts"
	etcdClient := EtcdClient{etcd, prefix}
	serial := "1"
	diskPath := "exists-path"
	key := "aaa"
	crypt := sabakanCrypt{Path: diskPath, Key: key}
	val, _ := json.Marshal(crypt)
	etcd.Put(context.Background(), path.Join(prefix, "crypts", serial, diskPath), string(val))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", path.Join("/api/v1/crypts", serial, diskPath), nil)
	r = mux.SetURLVars(r, map[string]string{"serial": serial, "path": diskPath})

	etcdClient.handleGetCrypts(w, r)

	resp := w.Result()
	var sut sabakanCrypt
	json.NewDecoder(resp.Body).Decode(&sut)
	if resp.StatusCode != 200 {
		t.Fatal("expected: 200, actual:", resp.StatusCode)
	}
	if sut != crypt {
		t.Fatal("invalid response body, ", sut)
	}
}

func TestHandleGetCryptsNotFound(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	etcdClient := EtcdClient{etcd, *flagEtcdPrefix + "/TestHandleGetCryptsNotFound"}
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
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := *flagEtcdPrefix + "/TestHandlePostCrypts"
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
	var responsedCrypt sabakanCrypt
	var savedCrypt sabakanCrypt
	json.NewDecoder(resp.Body).Decode(&responsedCrypt)
	etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyCrypts, serial, diskPath))
	json.Unmarshal(etcdResp.Kvs[0].Value, &savedCrypt)
	if resp.StatusCode != 201 {
		t.Fatal("expected: 201, actual:", resp.StatusCode)
	}
	if responsedCrypt != crypt {
		t.Fatal("invalid response body, ", responsedCrypt)
	}
	if savedCrypt != crypt {
		t.Fatal("saved entity is invalid, ", savedCrypt)
	}
}

func TestHandlePostCryptsInvalidBody(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	etcdClient := EtcdClient{etcd, *flagEtcdPrefix + "/TestHandlePostCryptsInvalidBody"}
	w := httptest.NewRecorder()
	invalidBody, _ := json.Marshal(&struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{"Taro", 1})
	r := httptest.NewRequest("POST", "/api/v1/crypts/1", bytes.NewBuffer(invalidBody))
	r = mux.SetURLVars(r, map[string]string{"serial": "1"})

	etcdClient.handlePostCrypts(w, r)

	resp := w.Result()
	if w.Result().StatusCode != 400 {
		t.Fatal("expected: 400, actual:", resp.StatusCode)
	}
}

func TestHandleDeleteCrypts(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := *flagEtcdPrefix + "/TestHandleDeleteCrypts"
	etcdClient := EtcdClient{etcd, prefix}
	expectedResponse := deleteResponse{}
	serial := "1"
	key := "aaa"
	for i := 0; i < 5; i++ {
		diskPath := "path" + strconv.Itoa(i)
		crypt := sabakanCrypt{Path: diskPath, Key: key}
		val, _ := json.Marshal(crypt)
		target := path.Join(prefix, EtcdKeyCrypts, serial, diskPath)
		etcd.Put(context.Background(), target, string(val))
		expectedResponse = append(expectedResponse, deletePath{target})
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", path.Join("/api/v1/crypts", serial), nil)
	r = mux.SetURLVars(r, map[string]string{"serial": serial})

	etcdClient.handleDeleteCrypts(w, r)

	resp := w.Result()
	var sut deleteResponse
	json.NewDecoder(resp.Body).Decode(&sut)
	etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyCrypts, serial), clientv3.WithPrefix())
	if resp.StatusCode != 200 {
		t.Fatal("expected: 200, actual:", resp.StatusCode)
	}
	if etcdResp.Count != 0 {
		t.Fatal("expected: 0, actual:", etcdResp.Count)
	}
	if !reflect.DeepEqual(sut, expectedResponse) {
		t.Fatal("unexpected response:", sut)
	}
}

func TestHandleDeleteCryptsNotFound(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	etcdClient := EtcdClient{etcd, *flagEtcdPrefix + "/TestHandleDeleteCryptsNotFound"}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/v1/crypts", nil)
	r = mux.SetURLVars(r, map[string]string{"serial": "non-exists-serial"})

	etcdClient.handleDeleteCrypts(w, r)

	resp := w.Result()
	if resp.StatusCode != 404 {
		t.Fatal("expected: 404, actual:", resp.StatusCode)
	}
}
