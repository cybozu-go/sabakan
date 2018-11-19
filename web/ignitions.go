package web

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/well"
)

func (s Server) handleIgnitions(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/boot/ignitions/"):], "/")
	if len(params) != 2 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	serial := params[0]
	id := params[1]

	if len(serial) == 0 || len(id) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}
	if r.Method != "GET" {
		renderError(r.Context(), w, APIErrBadMethod)
		return
	}

	s.serveIgnition(w, r, id, serial)
}

func (s Server) handleIgnitionTemplates(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/ignitions/"):], "/")

	if r.Method == "GET" && len(params) == 1 {
		role := params[0]
		if !sabakan.IsValidRole(role) {
			renderError(r.Context(), w, APIErrBadRequest)
			return
		}
		s.handleIgnitionTemplateIndexGet(w, r, role)
	} else if r.Method == "GET" && len(params) == 2 {
		role := params[0]
		id := params[1]
		if !sabakan.IsValidRole(role) {
			renderError(r.Context(), w, APIErrBadRequest)
			return
		}
		s.handleIgnitionTemplatesGet(w, r, role, id)
	} else if r.Method == "POST" && len(params) == 1 {
		role := params[0]
		if !sabakan.IsValidRole(role) {
			renderError(r.Context(), w, APIErrBadRequest)
			return
		}
		s.handleIgnitionTemplatesPost(w, r, role)
	} else if r.Method == "DELETE" && len(params) == 2 {
		role := params[0]
		id := params[1]
		if !sabakan.IsValidRole(role) || len(id) == 0 {
			renderError(r.Context(), w, APIErrBadRequest)
			return
		}
		s.handleIgnitionTemplatesDelete(w, r, role, id)
	} else {
		renderError(r.Context(), w, APIErrBadRequest)
	}
}

func (s Server) handleIgnitionTemplateIndexGet(w http.ResponseWriter, r *http.Request, role string) {
	metadata, err := s.Model.Ignition.GetTemplateMetadataList(r.Context(), role)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	renderJSON(w, metadata, http.StatusOK)
}

func (s Server) handleIgnitionTemplatesGet(w http.ResponseWriter, r *http.Request, role string, id string) {
	ign, err := s.Model.Ignition.GetTemplate(r.Context(), role, id)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write([]byte(ign))
	if err != nil {
		fields := well.FieldsFromContext(r.Context())
		fields[log.FnError] = err.Error()
		log.Error("failed to write response for GET /ignitions", fields)
	}
}

func (s Server) handleIgnitionTemplatesPost(w http.ResponseWriter, r *http.Request, role string) {
	// 1MB is maximum ignition template size
	body, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 1073741824))
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	ipam, err := s.Model.IPAM.GetConfig()
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	metadata := make(map[string]string)
	for k, v := range r.Header {
		optionHeaderPrefix := "x-sabakan-ignitions-"
		if strings.HasPrefix(strings.ToLower(k), optionHeaderPrefix) {
			key := strings.ToLower(k[len(optionHeaderPrefix):])
			if len(key) == 0 {
				continue
			}
			if !sabakan.IsValidLabelName(key) || key == "id" {
				renderError(r.Context(), w, BadRequest("invalid option key"+key))
				return
			}
			if !sabakan.IsValidLabelValue(v[0]) {
				renderError(r.Context(), w, BadRequest("invalid option value"+v[0]))
				return
			}
			metadata[key] = v[0]
		}
	}
	err = sabakan.ValidateIgnitionTemplate(string(body), metadata, ipam)
	if err != nil {
		renderError(r.Context(), w, BadRequest(err.Error()))
		return
	}

	id, err := s.Model.Ignition.PutTemplate(r.Context(), role, string(body), metadata)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	resp := make(map[string]interface{})
	resp["status"] = http.StatusCreated
	resp["role"] = role
	resp["id"] = id
	renderJSON(w, resp, http.StatusCreated)
}

func (s Server) handleIgnitionTemplatesDelete(w http.ResponseWriter, r *http.Request, role, id string) {
	err := s.Model.Ignition.DeleteTemplate(r.Context(), role, id)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s Server) serveIgnition(w http.ResponseWriter, r *http.Request, id, serial string) {
	m, err := s.Model.Machine.Get(r.Context(), serial)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	tmpl, err := s.Model.Ignition.GetTemplate(r.Context(), m.Spec.Role, id)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	meta, err := s.Model.Ignition.GetTemplateMetadata(r.Context(), m.Spec.Role, id)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	ign, err := sabakan.RenderIgnition(tmpl, meta, m, s.MyURL)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(ign))
	if err != nil {
		fields := well.FieldsFromContext(r.Context())
		fields[log.FnError] = err.Error()
		log.Error("failed to write response for GET /boot/ignitions", fields)
	}
}
