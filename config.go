package sabakan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	c *clientv3.Client
}

const (
	// EtcdKeyConfig is etcd key name for sabakan option
	EtcdKeyConfig = "/config"
	// EtcdKeyMachines is etcd key name for machines management
	EtcdKeyMachines = "/machines"

	// ErrorValueNotFound is an error message when a target value is not found
	ErrorValueNotFound = "Value not found"
	// ErrorMachinesExist is an error message when /machines key exists in etcd.
	ErrorMachinesExist = "Machines already exist"
	// ErrorValueAlreadyExists is an error message when a target value already exists
	ErrorValueAlreadyExists = "Value already exists"
)

// InitConfig is initialization of the sabakan API /config
func InitConfig(r *mux.Router, c *clientv3.Client) {
	e := &etcdClient{c}
	e.initConfigFunc(r)
}

func (e *etcdClient) initConfigFunc(r *mux.Router) {
	r.HandleFunc("/config", e.handleGetConfig).Methods("GET")
	r.HandleFunc("/config", e.handlePostConfig).Methods("POST")
}

func (e *etcdClient) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	resp, err := e.c.Get(r.Context(), EtcdKeyConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if resp == nil {
		http.Error(w, ErrorValueNotFound, http.StatusNotFound)
		return
	}
	if len(resp.Kvs) == 0 {
		http.Error(w, ErrorValueNotFound, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(resp.Kvs[0].Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (e *etcdClient) handlePostConfig(w http.ResponseWriter, r *http.Request) {
	resp, err := e.c.Get(r.Context(), EtcdKeyMachines, clientv3.WithPrefix())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if resp.Count != 0 {
		http.Error(w, ErrorMachinesExist, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var sc sabakanConfig

	fmt.Println(r.Body)
	b, _ := ioutil.ReadAll(r.Body)
	err = json.Unmarshal(b, &sc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Validation
	if sc.NodeIPv4Offset == "" {
		http.Error(w, "node-ipv4-offset: "+ErrorValueNotFound, http.StatusBadRequest)
		return
	}
	if sc.NodeRackShift == 0 {
		http.Error(w, "node-rack-shift: "+ErrorValueNotFound, http.StatusBadRequest)
		return
	}
	if sc.BMCIPv4Offset == "" {
		http.Error(w, "bmc-ipv4-offset: "+ErrorValueNotFound, http.StatusBadRequest)
		return
	}
	if sc.BMCRackShift == 0 {
		http.Error(w, "bmc-rack-shift: "+ErrorValueNotFound, http.StatusBadRequest)
		return
	}
	if sc.NodeIPPerNode == 0 {
		http.Error(w, "node-ip-per-node: "+ErrorValueNotFound, http.StatusBadRequest)
		return
	}
	if sc.BMCIPPerNode == 0 {
		http.Error(w, "bmc-ip-per-node: "+ErrorValueNotFound, http.StatusBadRequest)
		return
	}

	j, err := json.Marshal(sc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add config if /config doesn't exist.
	_, err = e.c.Txn(r.Context()).
		If(clientv3.Compare(clientv3.CreateRevision(EtcdKeyConfig), "=", 0)).
		Then(clientv3.OpPut(EtcdKeyConfig, string(j))).
		Else().
		Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add config if value of the /config is not same.
	_, err = e.c.Txn(r.Context()).
		If(clientv3.Compare(clientv3.Value(EtcdKeyConfig), "!=", string(j))).
		Then(clientv3.OpPut(EtcdKeyConfig, string(j))).
		Else().
		Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
