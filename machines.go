package sabakan

import (
	"encoding/json"
	"net/http"
)

type MachineJson struct {
	Serial           string                    `json:"serial"`
	Product          string                    `json:"product"`
	Datacenter       string                    `json:"datacenter"`
	Rack             *uint32                   `json:"rack"`
	NodeNumberOfRack *uint32                   `json:"node-number-of-rack"`
	Role             string                    `json:"role"`
	Network          map[string]MachineNetwork `json:"network"`
	BMC              MachineBMC                `json:"bmc"`
}

// ToMachine creates *Machine.
func (mj *MachineJson) ToMachine() *Machine {
	return &Machine{
		Serial:           mj.Serial,
		Product:          mj.Product,
		Datacenter:       mj.Datacenter,
		Rack:             *mj.Rack,
		NodeNumberOfRack: *mj.NodeNumberOfRack,
		Role:             mj.Role,
		Network:          mj.Network,
		BMC:              mj.BMC,
	}
}

// Machine is a machine struct
type Machine struct {
	Serial           string
	Product          string
	Datacenter       string
	Rack             uint32
	NodeNumberOfRack uint32
	Role             string
	Network          map[string]MachineNetwork
	BMC              MachineBMC
}

// ToJSON creates *MachineJson.
func (m *Machine) ToJSON() *MachineJson {
	rack := m.Rack
	num := m.NodeNumberOfRack
	return &MachineJson{
		Serial:           m.Serial,
		Product:          m.Product,
		Datacenter:       m.Datacenter,
		Rack:             &rack,
		NodeNumberOfRack: &num,
		Role:             m.Role,
		Network:          m.Network,
		BMC:              m.BMC,
	}
}

// MachineNetwork is a network interface struct for Machine
type MachineNetwork struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
	Mac  string   `json:"mac"`
}

func (n MachineNetwork) hasIPv4(ipv4 string) bool {
	for _, t := range n.IPv4 {
		if t == ipv4 {
			return true
		}
	}
	return false
}

func (n MachineNetwork) hasIPv6(ipv6 string) bool {
	for _, t := range n.IPv6 {
		if t == ipv6 {
			return true
		}
	}
	return false
}

// MachineBMC is a bmc interface struct for Machine
type MachineBMC struct {
	IPv4 []string `json:"ipv4"`
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
		machines[i] = mc.ToMachine()
	}

	err = s.Model.Machine.Register(r.Context(), machines)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getMachinesQuery(r *http.Request) *Query {
	var q Query
	vals := r.URL.Query()
	q.Serial = vals.Get("serial")
	q.Product = vals.Get("product")
	q.Datacenter = vals.Get("datacenter")
	q.Rack = vals.Get("rack")
	q.Role = vals.Get("role")
	q.IPv4 = vals.Get("ipv4")
	q.IPv6 = vals.Get("ipv6")
	return q
}

func (s Server) handleMachinesGet(w http.ResponseWriter, r *http.Request) {
	q := getMachinesQuery(r)

	machines, err := s.Model.Machine.Query(r.Context(), q)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	if len(machines) == 0 {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	j := make([]*MachineJson, len(machines))
	for i, m := range machines {
		j[i] = m.ToJSON()
	}

	renderJSON(w, j, http.StatusOK)
}
