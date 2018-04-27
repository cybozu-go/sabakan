package sabakan

import (
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/gorilla/mux"
)

// InitCrypts initialize the handle functions for crypts
func InitCrypts(r *mux.Router, e *EtcdClient) {
	e.initCryptsFunc(r)
}

func (e *EtcdClient) initCryptsFunc(r *mux.Router) {
	r.HandleFunc("/crypts/{serial}/{path}", e.handleGetCrypts).Methods("GET")
	r.HandleFunc("/crypts/{serial}/{path}", e.handlePutCrypts).Methods("PUT")
	r.HandleFunc("/crypts/{serial}", e.handleDeleteCrypts).Methods("DELETE")
}

func (e *EtcdClient) handleGetCrypts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	p := vars["path"]

	target := path.Join(e.Prefix, EtcdKeyCrypts, serial, p)
	resp, err := e.Client.Get(r.Context(), target)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}
	if resp.Count == 0 {
		renderError(w, errors.New(ErrorValueNotFound), http.StatusNotFound)
		return
	}

	ev := resp.Kvs[0]

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(ev.Value)))
	_, err = w.Write(ev.Value)
	if err != nil {
		fields := cmd.FieldsFromContext(r.Context())
		fields[log.FnError] = err.Error()
		log.Error("failed to write response for GET /crypts", fields)
	}
}

func (e *EtcdClient) handlePutCrypts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	p := vars["path"]

	keyData, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 4096))
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	if len(keyData) == 0 {
		renderError(w, errors.New("empty key data"), http.StatusBadRequest)
		return
	}

	target := path.Join(e.Prefix, EtcdKeyCrypts, serial, p)

	tresp, err := e.Client.Txn(r.Context()).
		// Prohibit overwriting
		If(clientv3util.KeyMissing(target)).
		Then(clientv3.OpPut(target, string(keyData))).
		Else().
		Commit()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}
	if !tresp.Succeeded {
		renderError(w, errors.New("sabakan prohibits overwriting crypt keys"), http.StatusConflict)
		return
	}

	resp := make(map[string]interface{})
	resp["status"] = http.StatusCreated
	resp["path"] = p

	renderJSON(w, resp, http.StatusCreated)
}

func (e *EtcdClient) handleDeleteCrypts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]
	target := path.Join(e.Prefix, EtcdKeyCrypts, serial) + "/"

	// DELETE
	dresp, err := e.Client.Delete(r.Context(), target,
		clientv3.WithPrefix(),
		clientv3.WithPrevKV(),
	)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}

	resp := make([]string, len(dresp.PrevKvs))
	for i, ev := range dresp.PrevKvs {
		resp[i] = string(ev.Key[len(target):])
	}
	renderJSON(w, resp, http.StatusOK)
}
