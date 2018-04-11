package sabakan

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
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
func InitCrypts(r *mux.Router, e *EtcdClient) {
	e.initCryptsFunc(r)
}

func (e *EtcdClient) initCryptsFunc(r *mux.Router) {
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

func (e *EtcdClient) handleGetCrypts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	diskPath := vars["path"]

	target := path.Join(e.Prefix, EtcdKeyCrypts, serial, diskPath)
	resp, err := e.Client.Get(r.Context(), target)
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

func (e *EtcdClient) handlePostCrypts(w http.ResponseWriter, r *http.Request) {
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

	if err := validatePostParams(received); err != nil {
		respError(w, err, http.StatusBadRequest)
		return
	}

	target := path.Join(e.Prefix, EtcdKeyCrypts, serial, diskPath)
	val, err := json.Marshal(sabakanCrypt{Path: diskPath, Key: key})

	tresp, err := e.Client.Txn(r.Context()).
		// Prohibit overwriting
		If(clientv3util.KeyMissing(target)).
		Then(clientv3.OpPut(target, string(val))).
		Else().
		Commit()
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if !tresp.Succeeded {
		respError(w, fmt.Errorf("transaction failed. sabakan prohibits overwriting crypts"), http.StatusInternalServerError)
		return
	}

	err = respWriter(w, sabakanCrypt{Path: string(diskPath), Key: string(key)}, http.StatusCreated)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
	}
}

func (e *EtcdClient) handleDeleteCrypts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	target := path.Join(e.Prefix, EtcdKeyCrypts, serial)

	// Confirm the targets exist
	gresp, err := e.Client.Get(r.Context(),
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
	_, err = e.Client.Delete(r.Context(),
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
