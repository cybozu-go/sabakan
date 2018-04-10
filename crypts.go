package sabakan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/cybozu-go/log"
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

	etcdClient struct {
		c *clientv3.Client
	}
)

// InitCrypts initialize the handle functions for crypts
func InitCrypts(r *mux.Router, c *clientv3.Client) {
	e := &etcdClient{c}
	e.initCryptsFunc(r)
}

func (e *etcdClient) initCryptsFunc(r *mux.Router) {
	r.HandleFunc("/{serial}/{path}", e.handleCryptsGet).Methods("GET")
	r.HandleFunc("/{serial}", e.handleCryptsPost).Methods("POST")
	r.HandleFunc("/{serial}", e.handleCryptsDelete).Methods("DELETE")
}

func respWriter(w http.ResponseWriter, data interface{}, status int) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
	return nil
}

func respError(w http.ResponseWriter, resperr error, status int) {
	out, err := json.Marshal(map[string]interface{}{
		"http_status_code": status,
		"error":            resperr.Error(),
	})
	if err != nil {
		log.Error(err.Error(), nil)
		return
	}

	log.Error(string(out), nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, string(out))
}

func makeDeleteResponse(gresp *clientv3.GetResponse, serial string) (deleteResponse, error) {
	entities := deleteResponse{}
	for _, ev := range gresp.Kvs {
		entities = append(entities, deleteResponseEntity{Path: string(ev.Key)})
	}
	return entities, nil
}

func (e *etcdClient) handleCryptsGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	path := vars["path"]

	target := fmt.Sprintf("/crypts/%v/%v", serial, path)
	resp, err := e.c.Get(r.Context(),
		target)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if resp.Count == 0 {
		respError(w, fmt.Errorf("target %v not found", target), http.StatusNotFound)
		return
	}

	ev := resp.Kvs[0]
	entity := &cryptEntity{Path: string(path), Key: string(ev.Value)}
	err = respWriter(w, entity, http.StatusOK)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}
}

func (e *etcdClient) handleCryptsPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	var receivedEntity cryptEntity
	b, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(b, &receivedEntity)
	path := receivedEntity.Path
	key := receivedEntity.Key
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if govalidator.IsNull(path) {
		respError(w, errors.New("`path` should not be empty"), http.StatusBadRequest)
		return
	}
	if govalidator.IsNull(key) {
		respError(w, errors.New("`key` should not be empty"), http.StatusBadRequest)
		return
	}

	//  Start mutex
	s, err := concurrency.NewSession(e.c)
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
	target := fmt.Sprintf("/crypts/%v/%v", serial, path)
	prev, err := e.c.Get(r.Context(), target)
	if err != nil {
		w.Write([]byte(err.Error() + "\n"))
		return
	}
	if prev.Count == 1 {
		respError(w, fmt.Errorf("target %v exists", target), http.StatusBadRequest)
		return
	}

	// Put crypts on etcd
	_, err = e.c.Txn(r.Context()).
		If(clientv3.Compare(clientv3.CreateRevision(target), "=", 0)).
		Then(clientv3.OpPut(target, key)).
		Else().
		Commit()
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	// Close mutex
	if err := m.Unlock(context.TODO()); err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	entity := &cryptEntity{Path: string(path), Key: string(key)}
	err = respWriter(w, entity, http.StatusCreated)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}
}

func (e *etcdClient) handleCryptsDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]

	// GET current crypts
	gresp, err := e.c.Get(r.Context(),
		fmt.Sprintf("/crypts/%v", serial),
		clientv3.WithPrefix())
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if len(gresp.Kvs) == 0 {
		respError(w, fmt.Errorf("target not found"), http.StatusNotFound)
		return
	}

	// DELETE
	dresp, err := e.c.Delete(r.Context(),
		fmt.Sprintf("/crypts/%v", serial),
		clientv3.WithPrefix())
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if dresp.Deleted <= 0 {
		respError(w, fmt.Errorf("failed to delete"), http.StatusInternalServerError)
		return
	}

	entities, err := makeDeleteResponse(gresp, serial)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}

	err = respWriter(w, entities, http.StatusOK)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}
}
