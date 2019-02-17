package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cybozu-go/sabakan"
)

const maxIgnitionTemplateSize = 2 * 1024 * 1024

func (s Server) handleIgnitionTemplates(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/ignitions/"):], "/")
	var role, id string

	// validate params
	if len(params) > 0 {
		role = params[0]
		if !sabakan.IsValidRole(role) {
			renderError(r.Context(), w, BadRequest("invalid role name: "+role))
			return
		}
	}
	if len(params) > 1 {
		id = params[1]
		if !sabakan.IsValidIgnitionID(id) {
			renderError(r.Context(), w, BadRequest("invalid ignition id: "+id))
			return
		}
	}

	switch {
	case r.Method == "GET" && len(params) == 1:
		s.handleIgnitionTemplateListIDs(w, r, role)
	case r.Method == "GET" && len(params) == 2:
		s.handleIgnitionTemplatesGet(w, r, role, id)
	case r.Method == "PUT" && len(params) == 2:
		s.handleIgnitionTemplatesPut(w, r, role, id)
	case r.Method == "DELETE" && len(params) == 2:
		s.handleIgnitionTemplatesDelete(w, r, role, id)
	default:
		renderError(r.Context(), w, APIErrBadRequest)
	}
}

func (s Server) handleIgnitionTemplateListIDs(w http.ResponseWriter, r *http.Request, role string) {
	ids, err := s.Model.Ignition.GetTemplateIDs(r.Context(), role)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	renderJSON(w, ids, http.StatusOK)
}

func (s Server) handleIgnitionTemplatesGet(w http.ResponseWriter, r *http.Request, role string, id string) {
	tmpl, err := s.Model.Ignition.GetTemplate(r.Context(), role, id)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	renderJSON(w, tmpl, http.StatusOK)
}

func (s Server) handleIgnitionTemplatesPut(w http.ResponseWriter, r *http.Request, role, id string) {
	tmpl := new(sabakan.IgnitionTemplate)
	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxIgnitionTemplateSize)).Decode(tmpl)
	if err != nil {
		renderError(r.Context(), w, BadRequest(fmt.Sprintf("invalid request body: %v", err)))
		return
	}

	err = s.validateIgnitionTemplate(tmpl)
	if err != nil {
		renderError(r.Context(), w, BadRequest(fmt.Sprintf("invalid template: %v", err)))
		return
	}

	err = s.Model.Ignition.PutTemplate(r.Context(), role, id, tmpl)
	if err == sabakan.ErrConflicted {
		renderError(r.Context(), w, APIErrConflict)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s Server) handleIgnitionTemplatesDelete(w http.ResponseWriter, r *http.Request, role, id string) {
	err := s.Model.Ignition.DeleteTemplate(r.Context(), role, id)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s Server) validateIgnitionTemplate(tmpl *sabakan.IgnitionTemplate) error {
	ipam, err := s.Model.IPAM.GetConfig()
	if err != nil {
		return err
	}
	mc := sabakan.NewMachine(sabakan.MachineSpec{
		Serial:      "1234abcd",
		Rack:        1,
		IndexInRack: 1,
	})
	ipam.GenerateIP(mc)

	_, err = s.renderIgnition(tmpl, mc)
	return err
}
