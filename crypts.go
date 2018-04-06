package sabakan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"io/ioutil"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gorilla/mux"
)

type cryptEntity struct {
	Path string `json:"path"`
	Key  string `json:"key"`
}

type etcdClient struct {
	c *clientv3.Client
}

func (e *etcdClient) initCryptsFunc(r *mux.Router) {
	r.HandleFunc("/{serial}/{path}", e.handleCryptsGet).Methods("GET")
	r.HandleFunc("/{serial}", e.handleCryptsPost).Methods("POST")
}

func InitCrypts(r *mux.Router, c *clientv3.Client) {
	e := &etcdClient{c}
	e.initCryptsFunc(r)
}

func (e *etcdClient) handleCryptsGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	path := vars["path"]

	resp, err := e.c.Get(r.Context(),
		fmt.Sprintf("/crypts/%v/%v", serial, path))
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	if len(resp.Kvs) == 0 {
		w.WriteHeader(404)
		return
	}
	if len(resp.Kvs) != 1 {
		w.WriteHeader(500)
		return
	}

	ev := resp.Kvs[0]
	entity := &cryptEntity{
		Path: string(path),
		Key:  string(ev.Value)}
	res, err := json.Marshal(entity)
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func (e *etcdClient) handleCryptsPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	var receivedEntity cryptEntity
	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &receivedEntity)
	path := receivedEntity.Path
	key := receivedEntity.Key

	s, err := concurrency.NewSession(e.c)
	defer s.Close()
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	m := concurrency.NewMutex(s, "/sabakan-post-crypts-lock/")
	if err := m.Lock(context.TODO()); err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}

	// Prohibit overwriting
	check, err := e.c.Get(r.Context(),
		fmt.Sprintf("/crypts/%v/%v", serial, path))
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	if len(check.Kvs) == 1 {
		w.WriteHeader(400)
		return
	}

	// Put crypts on etcd
	resp, err := e.c.Put(r.Context(),
		fmt.Sprintf("/crypts/%v/%v", serial, path),
		key)
	if err != nil && resp != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}

	// Confirm whether it was saved in etcd
	check, err = e.c.Get(r.Context(),
		fmt.Sprintf("/crypts/%v/%v", serial, path))
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	if len(check.Kvs) == 0 {
		w.WriteHeader(404)
		return
	}
	if len(check.Kvs) != 1 {
		w.WriteHeader(500)
		return
	}
	ev := check.Kvs[0]
	savedKey := string(ev.Value)
	if savedKey != key {
		w.WriteHeader(500)
		return
	}

	if err := m.Unlock(context.TODO()); err != nil {
		w.WriteHeader(500)
		return
	}

	responseEntity := &cryptEntity{
		Path: string(path),
		Key:  string(savedKey)}
	res, err := json.Marshal(responseEntity)
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(res)
}
