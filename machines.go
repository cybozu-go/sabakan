package sabakan

import (
	"encoding/json"
	"net/http"
	"reflect"
)

type MachineJson struct {
	Serial           string                    `json:"serial"`
	Product          string                    `json:"product"`
	Datacenter       string                    `json:"datacenter"`
	Rack             *uint32                   `json:"rack"`
	NodeNumberOfRack *uint32                   `json:"node-number-of-rack"`
	Role             string                    `json:"role"`
	Cluster          string                    `json:"cluster"`
	Network          map[string]MachineNetwork `json:"network"`
	BMC              MachineBMC                `json:"bmc"`
}

// Machine is a machine struct
type Machine struct {
	Serial           string
	Product          string
	Datacenter       string
	Rack             uint32
	NodeNumberOfRack uint32
	Role             string
	Cluster          string
	Network          map[string]MachineNetwork
	BMC              MachineBMC
}

// MachineNetwork is a network interface struct for Machine
type MachineNetwork struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
	Mac  string   `json:"mac"`
}

// MachineBMC is a bmc interface struct for Machine
type MachineBMC struct {
	IPv4 []string `json:"ipv4"`
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

func (s Server) handleMachines(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleMachinesGet(w, r)
		return
	case "POST":
		s.handleMachinesPost(w, r)
		return
	case "DELETE":
		s.handleMachinesDelete(w, r)
		return
	}

	renderError(r.Context(), w, APIErrBadRequest)
	return
}

func (s Server) handleMachinesPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var rmcs []MachineJson

	err := json.NewDecoder(r.Body).Decode(&rmcs)
	if err != nil {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	machines := make([]*Machine, len(rmcs))
	// Validation
	for i, mc := range rmcs {
		if mc.Serial == "" {
			renderError(r.Context(), w, BadRequest("serial is empty"))
			return
		}
		if mc.Product == "" {
			renderError(r.Context(), w, BadRequest("product is empty"))
			return
		}
		if mc.Datacenter == "" {
			renderError(r.Context(), w, BadRequest("datacenter is empty"))
			return
		}
		if mc.Rack == nil {
			renderError(r.Context(), w, BadRequest("rack is empty"))
			return
		}
		if mc.Role == "" {
			renderError(r.Context(), w, BadRequest("role is empty"))
			return
		}
		if mc.NodeNumberOfRack == nil {
			renderError(r.Context(), w, BadRequest("number of rack is empty"))
			return
		}
		if mc.Cluster == "" {
			renderError(r.Context(), w, BadRequest("cluster is empty"))
			return
		}
		machines[i] = &Machine{
			Serial:           mc.Serial,
			Product:          mc.Product,
			Datacenter:       mc.Datacenter,
			Rack:             *mc.Rack,
			NodeNumberOfRack: *mc.NodeNumberOfRack,
			Role:             mc.Role,
			Cluster:          mc.Cluster,
		}
	}

	err = s.Model.Machine.Register(r.Context(), machines)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getMachinesQuery(q *Query, r *http.Request) {
	q.Serial = r.URL.Query().Get("serial")
	q.Product = r.URL.Query().Get("product")
	q.Datacenter = r.URL.Query().Get("datacenter")
	q.Rack = r.URL.Query().Get("rack")
	q.Role = r.URL.Query().Get("role")
	q.Cluster = r.URL.Query().Get("cluster")
	q.IPv4 = r.URL.Query().Get("ipv4")
	q.IPv6 = r.URL.Query().Get("ipv6")
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

func (s Server) handleMachinesGet(w http.ResponseWriter, r *http.Request) {
	var q Query

	getMachinesQuery(&q, r)
	w.Header().Set("Content-Type", "application/json")
	result := []Machine{}

	if q.Serial != "" {
		mcs, err := GetMachinesBySerial(r.Context(), s, []string{q.Serial})
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		if len(mcs) != 0 {
			result = mcs
		}
		renderJSON(w, result, http.StatusOK)
		return
	}

	if q.IPv4 != "" {
		mcs, err := GetMachinesByIPv4(r.Context(), s, q.IPv4)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		if len(mcs) != 0 {
			result = mcs
		}
		renderJSON(w, result, http.StatusOK)
		return
	}

	if q.IPv6 != "" {
		mcs, err := GetMachinesByIPv4(r.Context(), s, q.IPv6)
		if err != nil {
			renderError(w, err, http.StatusInternalServerError)
			return
		}
		if len(mcs) != 0 {
			result = mcs
		}
		renderJSON(w, result, http.StatusOK)
		return
	}

	resultByQuery := map[string][]Machine{}
	queryCount := 0
	qelem := reflect.ValueOf(&q).Elem()

	for i := 0; i < qelem.NumField(); i++ {
		if qelem.Field(i).Interface().(string) != "" {
			queryCount++
		}
	}

	mi.mux.Lock()
	for i := 0; i < qelem.NumField(); i++ {
		qv := qelem.Field(i).Interface().(string)

		if q.Product != "" && mi.Product[qv] != nil {
			mcs, err := GetMachinesByProduct(r.Context(), s, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				mi.mux.Unlock()
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}

		if q.Datacenter != "" && mi.Datacenter[qv] != nil {
			mcs, err := GetMachinesByDatacenter(r.Context(), s, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				mi.mux.Unlock()
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}

		if q.Rack != "" && mi.Rack[qv] != nil {
			mcs, err := GetMachinesByRack(r.Context(), s, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				mi.mux.Unlock()
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}

		if q.Role != "" && mi.Role[qv] != nil {
			mcs, err := GetMachinesByRole(r.Context(), s, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				mi.mux.Unlock()
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}

		if q.Cluster != "" && mi.Cluster[qv] != nil {
			mcs, err := GetMachinesByCluster(r.Context(), s, qv)
			if err != nil {
				renderError(w, err, http.StatusInternalServerError)
				mi.mux.Unlock()
				return
			}
			if mcs != nil {
				resultByQuery[qv] = mcs
			}
			continue
		}
	}
	mi.mux.Unlock()

	if queryCount > 1 && queryCount == len(resultByQuery) {
		// Intersect machines of each query result
		for _, v := range resultByQuery {
			if len(result) == 0 {
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

	renderJSON(w, result, http.StatusOK)
}
