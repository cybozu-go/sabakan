package sabakan

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/etcd/clientv3"
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
	//r.HandleFunc("/{serial}/", e.handleCryptsPost).Methods("POST")
}

func InitCrypts(r *mux.Router, c *clientv3.Client) {
	e := &etcdClient{c}
	e.initCryptsFunc(r)
}

func (e *etcdClient) handleCryptsGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resp, err := e.c.Get(r.Context(),
		fmt.Sprintf("/crypts/%v/%v", vars["serial"], vars["path"]))
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}

	if len(resp.Kvs) == 0 {
		w.WriteHeader(404)
		return
	}

	for _, ev := range resp.Kvs {
		entity := &cryptEntity{
			Path: string(vars["path"]),
			Key:  string(ev.Value)}

		res, err := json.Marshal(entity)
		if err != nil {
			w.Write([]byte(err.Error() + "\n"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	}
}

//func (e *etcdClient) handleCryptsPost(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	resp, err := e.c.Put(r.Context(),
//		fmt.Sprintf("/crypts/%v/%v", vars["serial"], vars["path"]))
//
//}
