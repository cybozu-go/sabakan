package web

import (
	"encoding/json"
	"net/http"
	"strings"

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
	var specs []*sabakan.MachineSpec

	err := json.NewDecoder(r.Body).Decode(&specs)
	if err != nil {
		renderError(r.Context(), w, BadRequest(err.Error()))
		return
	}

	// Validation
	for _, m := range specs {
		if m.Serial == "" {
			renderError(r.Context(), w, BadRequest("serial is empty"))
			return
		}
		if m.Product == "" {
			renderError(r.Context(), w, BadRequest("product is empty"))
			return
		}
		if m.Datacenter == "" {
			renderError(r.Context(), w, BadRequest("datacenter is empty"))
			return
		}
		if !sabakan.IsValidRole(m.Role) {
			renderError(r.Context(), w, BadRequest("invalid role"))
			return
		}
		if m.BMC.Type == "" {
			renderError(r.Context(), w, BadRequest("BMC type is empty"))
			return
		}
		if !sabakan.IsValidBmcType(m.BMC.Type) {
			renderError(r.Context(), w, BadRequest("BMC type contains invalid character"))
			return
		}
		m.IPv4 = nil
		m.IPv6 = nil
	}
	machines := make([]*sabakan.Machine, len(specs))
	for i, spec := range specs {
		machines[i] = sabakan.NewMachine(*spec)
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

func getQueryMap(r *http.Request) sabakan.Query {
	q := make(sabakan.Query)
	vals := r.URL.Query()
	for k := range vals {
		q[k] = vals.Get(k)
	}
	return q
}

func (s Server) handleMachinesGet(w http.ResponseWriter, r *http.Request) {
	q := getQueryMap(r)

	machines, err := s.Model.Machine.Query(r.Context(), q)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	if len(machines) == 0 {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	j := make([]*sabakan.Machine, len(machines))
	for i, m := range machines {
		j[i] = m
	}

	renderJSON(w, j, http.StatusOK)
}

func (s Server) handleMachinesDelete(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/v1/machines/") {
		renderError(r.Context(), w, APIErrBadRequest)
	}
	serial := r.URL.Path[len("/api/v1/machines/"):]
	if len(serial) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
	}

	err := s.Model.Machine.Delete(r.Context(), serial)
	switch err {
	case nil:
		w.WriteHeader(http.StatusOK)
	case sabakan.ErrNotFound:
		renderError(r.Context(), w, APIErrNotFound)
	default:
		renderError(r.Context(), w, InternalServerError(err))
	}
}
