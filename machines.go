package sabakan

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"net"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/gorilla/mux"
)

type Machine struct {
	Serial           string                   `json:"serial"`
	Product          string                   `json:"product"`
	Datacenter       string                   `json:"datacenter"`
	Rack             uint32                   `json:"rack"`
	NodeNumberOfRack uint32                   `json:"node-number-of-rack"`
	Role             string                   `json:"role"`
	Cluster          string                   `json:"cluster"`
	Network          []map[string]interface{} `json:"network"`
	BMC              []map[string]interface{} `json:"bmc"`
}

const (
	ErrorValueIgnored  = "Value ignored"
	ErrorMachineExists = " already exists"
)

func InitMachines(r *mux.Router, c *clientv3.Client, p string) {
	e := &etcdClient{c, p}
	e.initMachinesFunc(r)
}

func (e *etcdClient) initMachinesFunc(r *mux.Router) {
	r.HandleFunc("/machines", e.handlePostMachines).Methods("POST")
	//r.HandleFunc("/machines", e.handlePutMachines).Methods("PUT")
	//r.HandleFunc("/{serial}", e.handleDeleteMachines).Methods("DELETE")
	//r.HandleFunc("/{serial}", e.handleGetMachines).Methods("GET")
}

func Offset2Uint32(offset string) uint32 {
	ip, _, _ := net.ParseCIDR(offset)
	var long uint32
	binary.Read(bytes.NewBuffer(ip.To4()), binary.BigEndian, &long)
	return long
}

func int642IP(n int64) string {
	b0 := strconv.FormatInt((n>>24)&0xff, 10)
	b1 := strconv.FormatInt((n>>16)&0xff, 10)
	b2 := strconv.FormatInt((n>>8)&0xff, 10)
	b3 := strconv.FormatInt((n & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

func (e *etcdClient) handlePostMachines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var mcs []Machine

	err := json.NewDecoder(r.Body).Decode(&mcs)
	if err != nil {
		respError(w, err, http.StatusBadRequest)
		return
	}

	// Validation
	for _, mc := range mcs {
		if mc.Serial == "" {
			respError(w, errors.New("serial: "+ErrorValueNotFound), http.StatusBadRequest)
			return
		}
		if mc.Product == "" {
			respError(w, errors.New("product: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Datacenter == "" {
			respError(w, errors.New("datacenter: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Rack == 0 {
			respError(w, errors.New("rack: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Role == "" {
			respError(w, errors.New("role: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.NodeNumberOfRack == 0 {
			respError(w, errors.New("node-number-of-rack: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Cluster == "" {
			respError(w, errors.New("cluster: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Network != nil {
			respError(w, errors.New("network: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.BMC != nil {
			respError(w, errors.New("bmc: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
	}

	for _, mc := range mcs {
		key := path.Join(e.prefix, EtcdKeyMachines, mc.Serial)
		resp, err := e.client.Get(r.Context(), key)
		if err != nil {
			respError(w, err, http.StatusInternalServerError)
			return
		}
		if resp.Count != 0 {
			respError(w, errors.New("Serial "+mc.Serial+ErrorMachineExists), http.StatusBadRequest)
			return
		}
	}

	// Generate IP addresses by sabakan config
	// net0 = node-ipv4-offset + (1 << node-rack-shift * 1 * rack-number) + node-number-of-a-rack
	// net1 = node-ipv4-offset + (1 << node-rack-shift * 2 * rack-number) + node-number-of-a-rack
	// net2 = node-ipv4-offset + (1 << node-rack-shift * 3 * rack-number) + node-number-of-a-rack
	// bmc  = bmc-ipv4-offset + (1 << bmc-rack-shift * 1 * rack-number) + node-number-of-a-rack
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

	var sc SabakanConfig
	err = json.Unmarshal(resp.Kvs[0].Value, &sc)
	if err != nil {
		respError(w, err, http.StatusBadRequest)
		return
	}

	for _, mc := range mcs {
		for i := 0; i < int(sc.NodeIPPerNode); i++ {
			uintNodeIPv4 := Offset2Uint32(sc.NodeIPv4Offset) + (uint32(1) << uint32(sc.NodeRackShift) * uint32(i+1) * mc.Rack) + mc.NodeNumberOfRack
			NodeIPv4 := int642IP(int64(uintNodeIPv4))
			ifname := fmt.Sprintf("net%d", i)
			net := map[string]interface{}{
				ifname: map[string]interface{}{
					"ipv4": []string{NodeIPv4},
					"ipv6": []string{},
					"mac":  "",
				},
			}
			mc.Network = append(mc.Network, net)
		}
		for i := 1; i <= int(sc.BMCIPPerNode); i++ {
			uintBMCIPv4 := Offset2Uint32(sc.BMCIPv4Offset) + (uint32(1) << uint32(sc.BMCRackShift) * uint32(i) * mc.Rack) + mc.NodeNumberOfRack
			BMCIPv4 := int642IP(int64(uintBMCIPv4))
			net := map[string]interface{}{
				"ipv4": []string{BMCIPv4},
			}
			mc.BMC = append(mc.BMC, net)
		}
	}

	// Add machines in a transaction
	txnIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, mc := range mcs {
		key = path.Join(e.prefix, EtcdKeyMachines, mc.Serial)
		txnIfOps = append(txnIfOps, clientv3util.KeyMissing(key))
		j, err := json.Marshal(mc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		txnThenOps = append(txnThenOps, clientv3.OpPut(key, string(j)))
	}
	_, err = e.client.Txn(r.Context()).
		If(
			txnIfOps...,
		).
		Then(
			txnThenOps...,
		).
		Else().
		Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
