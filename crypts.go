package sabakan

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gorilla/mux"
)

type (
	sabakanCrypt struct {
		Path string `json:"path"`
		Key  string `json:"key"`
	}

	deletePath struct {
		Path string `json:"path"`
	}

	deleteResponse []deletePath
)

// InitCrypts initialize the handle functions for crypts
func InitCrypts(r *mux.Router, c *clientv3.Client, p string) {
	e := &etcdClient{c, p}
	e.initCryptsFunc(r)
}

func (e *etcdClient) initCryptsFunc(r *mux.Router) {
	r.HandleFunc("/crypts/{serial}/{path}", e.handleGetCrypts).Methods("GET")
	r.HandleFunc("/crypts/{serial}", e.handlePostCrypts).Methods("POST")
	r.HandleFunc("/crypts/{serial}", e.handleDeleteCrypts).Methods("DELETE")
}

func makeDeleteResponse(gresp *clientv3.GetResponse) (deleteResponse, error) {
	dres := deleteResponse{}
	for _, ev := range gresp.Kvs {
		dres = append(dres, deletePath{Path: string(ev.Key)})
	}
	return dres, nil
}

func validatePostParams(received sabakanCrypt) error {
	diskPath := received.Path
	key := received.Key
	if len(diskPath) == 0 {
		return errors.New("`diskPath` should not be empty")
	}
	if len(key) == 0 {
		return errors.New("`key` should not be empty")
	}
	return nil
}

func (e *etcdClient) handleGetCrypts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	diskPath := vars["path"]

	target := path.Join(e.prefix, EtcdKeyCrypts, serial, diskPath)
	resp, err := e.client.Get(r.Context(), target)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if resp.Count == 0 {
		respError(w, fmt.Errorf(ErrorValueNotFound), http.StatusNotFound)
		return
	}

	ev := resp.Kvs[0]
	var responseBody sabakanCrypt
	err = json.Unmarshal(ev.Value, &responseBody)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	err = respWriter(w, responseBody, http.StatusOK)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}
}

func (e *etcdClient) handlePostCrypts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	var received sabakanCrypt
	err := json.NewDecoder(r.Body).Decode(&received)
	if err != nil {
		respError(w, err, http.StatusBadRequest)
		return
	}
	diskPath := received.Path
	key := received.Key

	// Validation
	if err := validatePostParams(received); err != nil {
		respError(w, err, http.StatusBadRequest)
		return
	}

	//  Start mutex
	s, err := concurrency.NewSession(e.client)
	defer s.Close()
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	m := concurrency.NewMutex(s, "/sabakan-post-crypts-lock/")
	if err := m.Lock(r.Context()); err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	// Prohibit overwriting
	target := path.Join(e.prefix, EtcdKeyCrypts, serial, diskPath)
	prev, err := e.client.Get(r.Context(), target)
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	if prev.Count >= 1 {
		respError(w, fmt.Errorf(ErrorCryptsExist), http.StatusBadRequest)
		return
	}

	// Put crypts on etcd
	crypt := sabakanCrypt{Path: diskPath, Key: key}
	val, err := json.Marshal(crypt)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if _, err := e.client.Put(r.Context(), target, string(val)); err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	// Close mutex
	if err := m.Unlock(r.Context()); err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	err = respWriter(w, sabakanCrypt{Path: string(diskPath), Key: string(key)}, http.StatusCreated)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}
}

func (e *etcdClient) handleDeleteCrypts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	target := path.Join(e.prefix, EtcdKeyCrypts, serial)

	// Confirm the targets exist
	gresp, err := e.client.Get(r.Context(),
		target,
		clientv3.WithPrefix())
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if gresp.Count == 0 {
		respError(w, fmt.Errorf(ErrorValueNotFound), http.StatusNotFound)
		return
	}

	// DELETE
	_, err = e.client.Delete(r.Context(),
		target,
		clientv3.WithPrefix())
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	dresp, err := makeDeleteResponse(gresp)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	err = respWriter(w, dresp, http.StatusOK)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}
}
