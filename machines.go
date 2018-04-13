package sabakan

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path"
	"reflect"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/netutil"
	"github.com/gorilla/mux"
)

// Machine is a machine struct
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

// Query is an URL query
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

// InitMachines is initilization of the sabakan API /machines
func InitMachines(r *mux.Router, e *EtcdClient) {
	e.initMachinesFunc(r)
}

func (e *EtcdClient) initMachinesFunc(r *mux.Router) {
	r.HandleFunc("/machines", e.handlePostMachines).Methods("POST")
	r.HandleFunc("/machines", e.handlePutMachines).Methods("PUT")
	r.HandleFunc("/machines", e.handleGetMachines).Methods("GET")
}

func offsetToInt(offset string) uint32 {
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

func generateIP(mc Machine, e *EtcdClient, r *http.Request) (Machine, int, error) {
	/*
		Generate IP addresses by sabakan config
		Example:
			net0 = node-ipv4-offset + (1 << node-rack-shift * 1 * rack-number) + node-number-of-a-rack
			net1 = node-ipv4-offset + (1 << node-rack-shift * 2 * rack-number) + node-number-of-a-rack
			net2 = node-ipv4-offset + (1 << node-rack-shift * 3 * rack-number) + node-number-of-a-rack
			bmc  = bmc-ipv4-offset + (1 << bmc-rack-shift * 1 * rack-number) + node-number-of-a-rack
	*/
	key := path.Join(e.Prefix, EtcdKeyConfig)
	resp, err := e.Client.Get(r.Context(), key)
	if err != nil {
		return Machine{}, http.StatusInternalServerError, err
	}
	if resp == nil {
		return Machine{}, http.StatusNotFound, fmt.Errorf(ErrorValueNotFound)
	}
	if len(resp.Kvs) == 0 {
		return Machine{}, http.StatusNotFound, fmt.Errorf(ErrorValueNotFound)
	}

	var sc Config
	err = json.Unmarshal(resp.Kvs[0].Value, &sc)
	if err != nil {
		return Machine{}, http.StatusBadRequest, err
	}

	for i := 0; i < int(sc.NodeIPPerNode); i++ {
		uintip := offsetToInt(sc.NodeIPv4Offset) + (uint32(1) << uint32(sc.NodeRackShift) * uint32(i+1) * mc.Rack) + mc.NodeNumberOfRack
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
		uintip := offsetToInt(sc.BMCIPv4Offset) + (uint32(1) << uint32(sc.BMCRackShift) * uint32(i+1) * mc.Rack) + mc.NodeNumberOfRack
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
	}, http.StatusOK, nil
}

func (e *EtcdClient) handlePostMachines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var rmcs []Machine

	err := json.NewDecoder(r.Body).Decode(&rmcs)
	if err != nil {
		renderError(w, err, http.StatusBadRequest)
		return
	}

	// Validation
	for _, mc := range rmcs {
		if mc.Serial == "" {
			renderError(w, fmt.Errorf("serial: "+ErrorValueNotFound), http.StatusBadRequest)
			return
		}
		if mc.Product == "" {
			renderError(w, fmt.Errorf("product: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Datacenter == "" {
			renderError(w, fmt.Errorf("datacenter: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Rack == 0 {
			renderError(w, fmt.Errorf("rack: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Role == "" {
			renderError(w, fmt.Errorf("role: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.NodeNumberOfRack == 0 {
			renderError(w, fmt.Errorf("node-number-of-rack: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Cluster == "" {
			renderError(w, fmt.Errorf("cluster: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.Network != nil {
			renderError(w, fmt.Errorf("network: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
		if mc.BMC != nil {
			renderError(w, fmt.Errorf("bmc: "+ErrorValueNotFound+" in the serial "+mc.Serial), http.StatusBadRequest)
			return
		}
	}

	for _, mc := range rmcs {
		key := path.Join(e.Prefix, EtcdKeyMachines, mc.Serial)
		resp, err := e.Client.Get(r.Context(), key)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		if resp.Count != 0 {
			renderError(w, fmt.Errorf("serial: "+mc.Serial+" "+ErrorMachineExists), http.StatusBadRequest)
			return
		}
	}

	wmcs := make([]Machine, len(rmcs))
	for i, rmc := range rmcs {
		status := 0
		wmcs[i], status, err = generateIP(rmc, e, r)
		if err != nil {
			renderError(w, err, status)
		}
	}

	// Put machines into etcd
	txnIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, wmc := range wmcs {
		key := path.Join(e.Prefix, EtcdKeyMachines, wmc.Serial)
		txnIfOps = append(txnIfOps, clientv3util.KeyMissing(key))
		j, err := json.Marshal(wmc)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		txnThenOps = append(txnThenOps, clientv3.OpPut(key, string(j)))
	}
	tresp, err := e.Client.Txn(r.Context()).
		If(
			txnIfOps...,
		).
		Then(
			txnThenOps...,
		).
		Else().
		Commit()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}
	if !tresp.Succeeded {
		renderError(w, fmt.Errorf(ErrorEtcdTxnFailed), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (e *EtcdClient) handlePutMachines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var rmcs []Machine

	err := json.NewDecoder(r.Body).Decode(&rmcs)
	if err != nil {
		renderError(w, err, http.StatusBadRequest)
		return
	}

	// Validation
	for _, mc := range rmcs {
		if mc.Serial == "" {
			renderError(w, fmt.Errorf("serial: "+ErrorValueNotFound), http.StatusBadRequest)
			return
		}
	}

	// Update []Machine
	wmcs := make([]Machine, len(rmcs))
	for i, rmc := range rmcs {
		key := path.Join(e.Prefix, EtcdKeyMachines, rmc.Serial)
		resp, err := e.Client.Get(r.Context(), key)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		if resp.Count == 0 {
			renderError(w, fmt.Errorf("Serial "+rmc.Serial+" "+ErrorMachineNotExists), http.StatusNotFound)
			return
		}

		var emc Machine
		err = json.Unmarshal(resp.Kvs[0].Value, &emc)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
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
		wmcs[i], status, err = generateIP(wmcs[i], e, r)
		if err != nil {
			renderError(w, err, status)
		}
	}

	// Put machines into etcd
	txnIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, wmc := range wmcs {
		key := path.Join(e.Prefix, EtcdKeyMachines, wmc.Serial)
		txnIfOps = append(txnIfOps, clientv3util.KeyExists(key))
		j, err := json.Marshal(wmc)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		txnThenOps = append(txnThenOps, clientv3.OpPut(key, string(j)))
	}
	tresp, err := e.Client.Txn(r.Context()).
		If(
			txnIfOps...,
		).
		Then(
			txnThenOps...,
		).
		Else().
		Commit()
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}
	if !tresp.Succeeded {
		renderError(w, fmt.Errorf(ErrorEtcdTxnFailed), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getMachinesQuery(q *Query, r *http.Request) error {
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

func intersectMachines(mcs1 []Machine, mcs2 []Machine) []Machine {
	var newmcs []Machine
	for _, mc1 := range mcs1 {
		for _, mc2 := range mcs2 {
			if mc1.Serial == mc2.Serial {
				newmcs = append(newmcs, mc2)
			}
		}
	}
	return newmcs
}

func (e *EtcdClient) handleGetMachines(w http.ResponseWriter, r *http.Request) {
	var q Query

	if err := getMachinesQuery(&q, r); err != nil {
		renderError(w, err, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if q.Serial != "" {
		j, err := GetMachineBySerial(e, r, q.Serial)
		if err != nil {
			renderError(w, err, http.StatusNotFound)
			return
		}
		var responseBody Machine
		err = json.Unmarshal(j, &responseBody)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
		}
		err = renderJSON(w, responseBody, http.StatusOK)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		return
	}

	if q.IPv4 != "" {
		j, err := GetMachineByIPv4(e, r, q.IPv4)
		if err != nil {
			renderError(w, err, http.StatusNotFound)
			return
		}
		var responseBody Machine
		err = json.Unmarshal(j, &responseBody)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
		}
		err = renderJSON(w, responseBody, http.StatusOK)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		return
	}

	if q.IPv6 != "" {
		j, err := GetMachineByIPv6(e, r, q.IPv4)
		if err != nil {
			renderError(w, err, http.StatusNotFound)
			return
		}
		var responseBody Machine
		err = json.Unmarshal(j, &responseBody)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
		}
		err = renderJSON(w, responseBody, http.StatusOK)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		return
	}

	var resultByQuery map[string][]Machine
	resultByQuery = map[string][]Machine{}
	queryCount := 0
	qelem := reflect.ValueOf(&q).Elem()

	for i := 0; i < qelem.NumField(); i++ {
		if qelem.Field(i).Interface().(string) != "" {
			queryCount++
		}
	}

	for i := 0; i < qelem.NumField(); i++ {
		qv := qelem.Field(i).Interface().(string)

		if q.Product != "" && MI.Product[qv] != nil {
			mcs, err := GetMachinesByProduct(e, r, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}

		if q.Datacenter != "" && MI.Datacenter[qv] != nil {
			mcs, err := GetMachinesByDatacenter(e, r, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}

		if q.Rack != "" && MI.Rack[qv] != nil {
			mcs, err := GetMachinesByRack(e, r, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}

		if q.Role != "" && MI.Role[qv] != nil {
			mcs, err := GetMachinesByRole(e, r, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}

		if q.Cluster != "" && MI.Cluster[qv] != nil {
			mcs, err := GetMachinesByCluster(e, r, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}
	}

	result := make([]Machine, 0)
	if queryCount > 1 && queryCount == len(resultByQuery) {
		// Intersect machines of each query result
		for _, v := range resultByQuery {
			if result == nil {
				result = v
				continue
			}
			result = intersectMachines(result, v)
		}
	}
	if queryCount == 1 && queryCount == len(resultByQuery) {
		for _, v := range resultByQuery {
			result = v
		}
	}

	err := renderJSON(w, result, http.StatusOK)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
		return
	}
}

// EtcdWatcher launch etcd client session to monitor changes to keys and update index
func EtcdWatcher(e EtcdConfig) {
	go func() {
		cfg := clientv3.Config{
			Endpoints: e.Servers,
		}
		c, err := clientv3.New(cfg)
		if err != nil {
			log.ErrorExit(err)
		}
		defer c.Close()

		key := path.Join(e.Prefix, EtcdKeyMachines)
		rch := c.Watch(context.TODO(), key, clientv3.WithPrefix(), clientv3.WithPrevKV())
		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.PUT && ev.PrevKv != nil {
					UpdateIndex(ev.PrevKv.Value, ev.Kv.Value)
				}
				if ev.Type == mvccpb.PUT && ev.PrevKv == nil {
					AddIndex(ev.Kv.Value)
				}
				if ev.Type == mvccpb.DELETE {
					DeleteIndex(ev.PrevKv.Value)
				}
			}
		}
	}()
}
