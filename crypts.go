package sabakan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"path"

	"github.com/asaskevich/govalidator"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gorilla/mux"
)

type (
	cryptEntity struct {
		Path string `json:"path"`
		Key  string `json:"key"`
	}

	deleteResponseEntity struct {
		Path string `json:"path"`
	}

	deleteResponse []deleteResponseEntity
)

// InitCrypts initialize the handle functions for crypts
func InitCrypts(r *mux.Router, c *clientv3.Client, p string) {
	e := &etcdClient{c, p}
	e.initCryptsFunc(r)
}

func (e *etcdClient) initCryptsFunc(r *mux.Router) {
	r.HandleFunc(EtcdKeyCrypts+"/{serial}/{path}", e.handleCryptsGet).Methods("GET")
	r.HandleFunc(EtcdKeyCrypts+"/{serial}", e.handleCryptsPost).Methods("POST")
	r.HandleFunc(EtcdKeyCrypts+"/{serial}", e.handleCryptsDelete).Methods("DELETE")
}

func makeDeleteResponse(gresp *clientv3.GetResponse) (deleteResponse, error) {
	entities := deleteResponse{}
	for _, ev := range gresp.Kvs {
		entities = append(entities, deleteResponseEntity{Path: string(ev.Key)})
	}
	return entities, nil
}

func validatePostParams(received cryptEntity) error {
	diskPath := received.Path
	key := received.Key
	if govalidator.IsNull(diskPath) {
		s := "`diskPath` should not be empty"
		return errors.New(s)
	}
	if govalidator.IsNull(key) {
		return errors.New("`key` should not be empty")
	}
	return nil
}

func (e *etcdClient) handleCryptsGet(w http.ResponseWriter, r *http.Request) {
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
	var responseBody cryptEntity
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

func (e *etcdClient) handleCryptsPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	var received cryptEntity
	err := json.NewDecoder(r.Body).Decode(&received)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
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
	if err := m.Lock(context.TODO()); err != nil {
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
	entity := cryptEntity{Path: diskPath, Key: key}
	val, err := json.Marshal(entity)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if _, err := e.client.Put(r.Context(), target, string(val)); err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	// Close mutex
	if err := m.Unlock(context.TODO()); err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	err = respWriter(w, cryptEntity{Path: string(diskPath), Key: string(key)}, http.StatusCreated)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}
}

func (e *etcdClient) handleCryptsDelete(w http.ResponseWriter, r *http.Request) {
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

	entities, err := makeDeleteResponse(gresp)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	err = respWriter(w, entities, http.StatusOK)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}
}
