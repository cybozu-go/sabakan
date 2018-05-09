package web

import (
	"encoding/json"
	"net/http"

	"github.com/cybozu-go/sabakan"
)

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

	renderError(r.Context(), w, APIErrBadMethod)
}

func (s Server) handleMachinesPost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var rmcs []sabakan.MachineJSON

	err := json.NewDecoder(r.Body).Decode(&rmcs)
	if err != nil {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	machines := make([]*sabakan.Machine, len(rmcs))
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
	switch err {
	case sabakan.ErrConflicted:
		renderError(r.Context(), w, APIErrConflict)
	case nil:
	default:
		renderError(r.Context(), w, InternalServerError(err))
	}

	w.WriteHeader(http.StatusCreated)
}

func getMachinesQuery(r *http.Request) *sabakan.Query {
	var q sabakan.Query
	vals := r.URL.Query()
	q.Serial = vals.Get("serial")
	q.Product = vals.Get("product")
	q.Datacenter = vals.Get("datacenter")
	q.Rack = vals.Get("rack")
	q.Role = vals.Get("role")
	q.IPv4 = vals.Get("ipv4")
	q.IPv6 = vals.Get("ipv6")
	return &q
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

	j := make([]*sabakan.MachineJSON, len(machines))
	for i, m := range machines {
		j[i] = m.ToJSON()
	}

	renderJSON(w, j, http.StatusOK)
}

func (s Server) handleMachinesDelete(w http.ResponseWriter, r *http.Request) {
	//TODO
}
