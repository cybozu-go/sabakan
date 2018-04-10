package sabakan

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/gorilla/mux"
)

// SabakanConfig is structure of the sabakan option
type sabakanConfig struct {
	NodeIPv4Offset string `json:"node-ipv4-offset"`
	NodeRackShift  uint   `json:"node-rack-shift"`
	BMCIPv4Offset  string `json:"bmc-ipv4-offset"`
	BMCRackShift   uint   `json:"bmc-rack-shift"`
	NodeIPPerNode  uint   `json:"node-ip-per-node"`
	BMCIPPerNode   uint   `json:"bmc-ip-per-node"`
}

// etcdClient is etcd3 client object
type etcdClient struct {
	client *clientv3.Client
	prefix string
}

const (
	// EtcdKeyConfig is etcd key name for sabakan option
	EtcdKeyConfig = "/config"
	// EtcdKeyMachines is etcd key name for machines management
	EtcdKeyMachines = "/machines"
	// EtcdKeyCrypts is etcd key name for crypts management
	EtcdKeyCrypts = "/crypts"

	// ErrorValueNotFound is an error message when a target value is not found
	ErrorValueNotFound = "value not found"
	// ErrorMachinesExist is an error message when /machines key exists in etcd.
	ErrorMachinesExist = "machines already exist"
	// ErrorCryptsExist is an error message when /crypts key exists in etcd.
	ErrorCryptsExist = "crypts already exist"
)

// InitConfig is initialization of the sabakan API /config
func InitConfig(r *mux.Router, c *clientv3.Client, p string) {
	e := &etcdClient{c, p}
	e.initConfigFunc(r)
}

func (e *etcdClient) initConfigFunc(r *mux.Router) {
	r.HandleFunc("/config", e.handleGetConfig).Methods("GET")
	r.HandleFunc("/config", e.handlePostConfig).Methods("POST")
}

func (e *etcdClient) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	key := path.Join(e.prefix, EtcdKeyConfig)
	resp, err := e.client.Get(r.Context(), key)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if resp == nil {
		respError(w, errors.New(ErrorValueNotFound), http.StatusNotFound)
		return
	}
	if len(resp.Kvs) == 0 {
		respError(w, errors.New(ErrorValueNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp.Kvs[0].Value)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
}

func (e *etcdClient) handlePostConfig(w http.ResponseWriter, r *http.Request) {
	key := path.Join(e.prefix, EtcdKeyMachines)
	resp, err := e.client.Get(r.Context(), key, clientv3.WithPrefix())
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}
	if resp.Count != 0 {
		respError(w, errors.New(ErrorMachinesExist), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var sc sabakanConfig

	err = json.NewDecoder(r.Body).Decode(&sc)
	if err != nil {
		respError(w, err, http.StatusBadRequest)
		return
	}
	// Validation
	if sc.NodeIPv4Offset == "" {
		respError(w, errors.New("node-ipv4-offset: "+ErrorValueNotFound), http.StatusBadRequest)
		return
	}
	if sc.NodeRackShift == 0 {
		respError(w, errors.New("node-rack-shift: "+ErrorValueNotFound), http.StatusBadRequest)
		return
	}
	if sc.BMCIPv4Offset == "" {
		respError(w, errors.New("bmc-ipv4-offset: "+ErrorValueNotFound), http.StatusBadRequest)
		return
	}
	if sc.BMCRackShift == 0 {
		respError(w, errors.New("bmc-rack-shift: "+ErrorValueNotFound), http.StatusBadRequest)
		return
	}
	if sc.NodeIPPerNode == 0 {
		respError(w, errors.New("node-ip-per-node: "+ErrorValueNotFound), http.StatusBadRequest)
		return
	}
	if sc.BMCIPPerNode == 0 {
		respError(w, errors.New("bmc-ip-per-node: "+ErrorValueNotFound), http.StatusBadRequest)
		return
	}

	j, err := json.Marshal(sc)
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	// Put config
	key = path.Join(e.prefix, EtcdKeyConfig)
	_, err = e.client.Put(r.Context(), key, string(j))
	if err != nil {
		respError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
