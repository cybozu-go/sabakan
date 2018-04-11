package sabakan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"

	"net"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/netutil"
	"github.com/gorilla/mux"
)

type Machine struct {
	Serial           string                 `json:"serial"`
	Product          string                 `json:"product"`
	Datacenter       string                 `json:"datacenter"`
	Rack             uint32                 `json:"rack"`
	NodeNumberOfRack uint32                 `json:"node-number-of-rack"`
	Role             string                 `json:"role"`
	Cluster          string                 `json:"cluster"`
	Network          map[string]interface{} `json:"network"`
	BMC              map[string]interface{} `json:"bmc"`
}

type Query struct {
	Serial     string
	Product    string
	Datacenter string
	Rack       string
	Role       string
	Cluster    string
	IPv4       string
	IPv6       string
}

const (
	ErrorValueIgnored    = "value ignored"
	ErrorMachineExists   = "already exists"
	ErrorMachineNotFound = "machine not found"
)

func InitMachines(r *mux.Router, c *clientv3.Client, p string) {
	e := &etcdClient{c, p}
	e.initMachinesFunc(r)
}

func (e *etcdClient) initMachinesFunc(r *mux.Router) {
	r.HandleFunc("/machines", e.handlePostMachines).Methods("POST")
	r.HandleFunc("/machines", e.handlePutMachines).Methods("PUT")
	//r.HandleFunc("/machines", e.handleDeleteMachines).Methods("DELETE")
	r.HandleFunc("/machines", e.handleGetMachines).Methods("GET")
}

func OffsetToInt(offset string) uint32 {
	ip, _, _ := net.ParseCIDR(offset)
	return netutil.IP4ToInt(ip)
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

func generateIP(mc Machine, e *etcdClient, r *http.Request) (Machine, error, int) {
	/*
		Generate IP addresses by sabakan config
		Example:
			net0 = node-ipv4-offset + (1 << node-rack-shift * 1 * rack-number) + node-number-of-a-rack
			net1 = node-ipv4-offset + (1 << node-rack-shift * 2 * rack-number) + node-number-of-a-rack
			net2 = node-ipv4-offset + (1 << node-rack-shift * 3 * rack-number) + node-number-of-a-rack
			bmc  = bmc-ipv4-offset + (1 << bmc-rack-shift * 1 * rack-number) + node-number-of-a-rack
	*/
	key := path.Join(e.prefix, EtcdKeyConfig)
	resp, err := e.client.Get(r.Context(), key)
	if err != nil {
		return Machine{}, err, http.StatusInternalServerError
	}
	if resp == nil {
		return Machine{}, fmt.Errorf(ErrorValueNotFound), http.StatusNotFound
	}
	if len(resp.Kvs) == 0 {
		return Machine{}, fmt.Errorf(ErrorValueNotFound), http.StatusNotFound
	}

	var sc sabakanConfig
	err = json.Unmarshal(resp.Kvs[0].Value, &sc)
	if err != nil {
		return Machine{}, err, http.StatusBadRequest
	}

	for i := 0; i < int(sc.NodeIPPerNode); i++ {
		uintip := OffsetToInt(sc.NodeIPv4Offset) + (uint32(1) << uint32(sc.NodeRackShift) * uint32(i+1) * mc.Rack) + mc.NodeNumberOfRack
		ip := netutil.IntToIP4(uintip)
		ifname := fmt.Sprintf("net%d", i)
		nif := map[string]interface{}{
			ifname: map[string]interface{}{
				"ipv4": []string{ip.String()},
				"ipv6": []string{},
				"mac":  "",
			},
		}
		mc.Network = mergeMaps(mc.Network, nif)
	}
	for i := 0; i < int(sc.BMCIPPerNode); i++ {
		uintip := OffsetToInt(sc.BMCIPv4Offset) + (uint32(1) << uint32(sc.BMCRackShift) * uint32(i+1) * mc.Rack) + mc.NodeNumberOfRack
		ip := netutil.IntToIP4(uintip)
		nif := map[string]interface{}{
			"ipv4": []string{ip.String()},
		}
		mc.BMC = mergeMaps(mc.BMC, nif)
	}

	return Machine{
		mc.Serial,
		mc.Product,
		mc.Datacenter,
		mc.Rack,
		mc.NodeNumberOfRack,
		mc.Role,
		mc.Cluster,
		mc.Network,
		mc.BMC,
	}, nil, http.StatusOK
}

func (e *etcdClient) handlePostMachines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var rmcs []Machine

	err := json.NewDecoder(r.Body).Decode(&rmcs)
	if err != nil {
		respError(w, err, http.StatusBadRequest)
		return
	}

	// Validation
	for _, mc := range rmcs {
		if mc.Serial == "" {
			respError(w, fmt.Errorf("serial: "+ErrorValueNotFound), http.StatusBadRequest)
			return
		}
		if mc.Product == "" {
			respError(w, fmt.Errorf("product: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Datacenter == "" {
			respError(w, fmt.Errorf("datacenter: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Rack == 0 {
			respError(w, fmt.Errorf("rack: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Role == "" {
			respError(w, fmt.Errorf("role: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.NodeNumberOfRack == 0 {
			respError(w, fmt.Errorf("node-number-of-rack: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Cluster == "" {
			respError(w, fmt.Errorf("cluster: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Network != nil {
			respError(w, fmt.Errorf("network: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.BMC != nil {
			respError(w, fmt.Errorf("bmc: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
	}

	for _, mc := range rmcs {
		key := path.Join(e.prefix, EtcdKeyMachines, mc.Serial)
		resp, err := e.client.Get(r.Context(), key)
		if err != nil {
			respError(w, err, http.StatusInternalServerError)
			return
		}
		if resp.Count != 0 {
			respError(w, fmt.Errorf("serial: "+mc.Serial+" "+ErrorMachineExists), http.StatusBadRequest)
			return
		}
	}

	wmcs := make([]Machine, len(rmcs))
	for i, rmc := range rmcs {
		status := 0
		wmcs[i], err, status = generateIP(rmc, e, r)
		if err != nil {
			respError(w, err, status)
		}
	}

	// Put machines into etcd
	txnIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, wmc := range wmcs {
		key := path.Join(e.prefix, EtcdKeyMachines, wmc.Serial)
		txnIfOps = append(txnIfOps, clientv3util.KeyMissing(key))
		j, err := json.Marshal(wmc)
		if err != nil {
			respError(w, err, http.StatusInternalServerError)
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
		respError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (e *etcdClient) handlePutMachines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var rmcs []Machine

	err := json.NewDecoder(r.Body).Decode(&rmcs)
	if err != nil {
		respError(w, err, http.StatusBadRequest)
		return
	}

	// Validation
	for _, mc := range rmcs {
		if mc.Serial == "" {
			respError(w, fmt.Errorf("serial: "+ErrorValueNotFound), http.StatusBadRequest)
			return
		}
	}

	// Update []Machine
	wmcs := make([]Machine, len(rmcs))
	for i, rmc := range rmcs {
		key := path.Join(e.prefix, EtcdKeyMachines, rmc.Serial)
		resp, err := e.client.Get(r.Context(), key)
		if err != nil {
			respError(w, err, http.StatusInternalServerError)
			return
		}
		if resp.Count == 0 {
			respError(w, fmt.Errorf("Serial "+rmc.Serial+ErrorMachineNotFound), http.StatusNotFound)
			return
		}

		var emc Machine
		err = json.Unmarshal(resp.Kvs[0].Value, &emc)
		if err != nil {
			respError(w, err, http.StatusInternalServerError)
			return
		}
		wmcs[i] = emc

		if rmc.Product != "" && emc.Product != rmc.Product {
			wmcs[i].Product = rmc.Product
		}
		if rmc.Datacenter != "" && emc.Datacenter != rmc.Datacenter {
			wmcs[i].Datacenter = rmc.Datacenter
		}
		if rmc.Rack != 0 && emc.Rack != rmc.Rack {
			wmcs[i].Rack = rmc.Rack
		}
		if rmc.Role != "" && emc.Role != rmc.Role {
			wmcs[i].Role = rmc.Role
		}
		if rmc.NodeNumberOfRack != 0 && emc.NodeNumberOfRack != rmc.NodeNumberOfRack {
			wmcs[i].NodeNumberOfRack = rmc.NodeNumberOfRack
		}
		if rmc.Cluster != "" && emc.Cluster != rmc.Cluster {
			wmcs[i].Cluster = rmc.Cluster
		}

		status := 0
		wmcs[i], err, status = generateIP(wmcs[i], e, r)
		if err != nil {
			respError(w, err, status)
		}
	}

	// Put machines into etcd
	txnIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, wmc := range wmcs {
		key := path.Join(e.prefix, EtcdKeyMachines, wmc.Serial)
		txnIfOps = append(txnIfOps, clientv3util.KeyExists(key))
		j, err := json.Marshal(wmc)
		if err != nil {
			respError(w, err, http.StatusInternalServerError)
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
		respError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleGetMachinesParam(q *Query, r *http.Request) error {
	q.Serial = r.URL.Query().Get("serial")
	q.Product = r.URL.Query().Get("product")
	q.Datacenter = r.URL.Query().Get("datacenter")
	q.Rack = r.URL.Query().Get("rack")
	q.Role = r.URL.Query().Get("role")
	q.Cluster = r.URL.Query().Get("cluster")
	q.IPv4 = r.URL.Query().Get("ipv4")
	q.IPv6 = r.URL.Query().Get("ipv6")

	return nil
}

func writeHandleGetMachines(w http.ResponseWriter, result []Machine) error {
	out, err := json.Marshal(result)
	if err != nil {
		return err
	}
	w.Write(out)
	w.WriteHeader(http.StatusOK)
	return nil
}

func appendMachinesIfNotExist(mcs []Machine, newmcs []Machine) []Machine {
	for _, newmc := range newmcs {
		found := false
		for _, mc := range mcs {
			if mc.Serial == newmc.Serial {
				found = true
			}
		}
		if mcs == nil || !found {
			mcs = append(mcs, newmc)
		}
	}
	return mcs
}

func (e *etcdClient) handleGetMachines(w http.ResponseWriter, r *http.Request) {
	var q Query

	if err := handleGetMachinesParam(&q, r); err != nil {
		respError(w, err, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if q.Serial != "" {
		j, err := GetMachineBySerial(e, r, q.Serial)
		if err != nil {
			respError(w, err, http.StatusNotFound)
			return
		}
		w.Write(j)
		w.WriteHeader(http.StatusOK)
		return
	}

	if q.IPv4 != "" {
		j, err := GetMachineByIPv4(e, r, q.IPv4)
		if err != nil {
			respError(w, err, http.StatusNotFound)
			return
		}
		w.Write(j)
		w.WriteHeader(http.StatusOK)
		return
	}
	if q.IPv6 != "" {
		j, err := GetMachineByIPv6(e, r, q.IPv4)
		if err != nil {
			respError(w, err, http.StatusNotFound)
			return
		}
		w.Write(j)
		w.WriteHeader(http.StatusOK)
		return
	}

	var result_t map[string][]Machine
	result_t = map[string][]Machine{}
	var result []Machine
	qelem := reflect.ValueOf(&q).Elem()
	//typeOfqelem := qelem.Type()

	for i := 0; i < qelem.NumField(); i++ {
		//fmt.Println(qelem.Field(i))
		//fmt.Println(typeOfqelem.Field(i).Name)
		qv := qelem.Field(i).Interface().(string)

		if q.Product != "" && MI.Product[qv] != nil {
			mcs, err := GetMachinesByProduct(e, r, qv)
			if err != nil {
				respError(w, err, http.StatusInternalServerError)
				return
			}

			result_t[qv] = mcs
			continue
		}

		if q.Datacenter != "" && MI.Datacenter[qv] != nil {
			mcs, err := GetMachinesByDatacenter(e, r, qv)
			if err != nil {
				respError(w, err, http.StatusInternalServerError)
				return
			}
			result_t[qv] = mcs
			continue
		}

		if q.Rack != "" && MI.Rack[qv] != nil {
			mcs, err := GetMachinesByRack(e, r, qv)
			if err != nil {
				respError(w, err, http.StatusInternalServerError)
				return
			}
			result = append(result, mcs...)
			continue
		}

		if q.Role != "" && MI.Role[qv] != nil {
			mcs, err := GetMachinesByRole(e, r, qv)
			if err != nil {
				respError(w, err, http.StatusInternalServerError)
				return
			}
			result = append(result, mcs...)
			continue
		}

		if q.Cluster != "" && MI.Cluster[qv] != nil {
			mcs, err := GetMachinesByCluster(e, r, qv)
			if err != nil {
				respError(w, err, http.StatusInternalServerError)
				return
			}
			result = append(result, mcs...)
			continue
		}
	}

	err := writeHandleGetMachines(w, result)
	if err != nil {
		respError(w, err, http.StatusNotFound)
		return
	}

}
