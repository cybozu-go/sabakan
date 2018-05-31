package web

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

func (s Server) handleIgnitions(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/boot/ignitions/"):], "/")
	if len(params) != 3 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	role := params[0]
	id := params[1]
	serial := params[2]

	if len(role) == 0 || len(id) == 0 || len(serial) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}
	if r.Method != "GET" {
		renderError(r.Context(), w, APIErrBadMethod)
		return
	}

	s.serveIgnition(w, r, role, id, serial)
}

func (s Server) handleIgnitionTemplates(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/ignitions/"):], "/")

	if r.Method == "GET" && len(params) == 1 {
		role := params[0]
		if len(role) == 0 {
			renderError(r.Context(), w, APIErrBadRequest)
			return
		}
		s.handleIgnitionTemplatesGet(w, r, role)
	} else if r.Method == "PUT" && len(params) == 1 {
		role := params[0]
		if len(role) == 0 {
			renderError(r.Context(), w, APIErrBadRequest)
			return
		}
		s.handleIgnitionTemplatesPut(w, r, role)
	} else if r.Method == "DELETE" && len(params) == 2 {
		role := params[0]
		id := params[1]
		if len(role) == 0 || len(id) == 0 {
			renderError(r.Context(), w, APIErrBadRequest)
			return
		}
		s.handleIgnitionTemplatesDelete(w, r, role, id)
	} else {
		renderError(r.Context(), w, APIErrBadRequest)
	}
}

func (s Server) handleIgnitionTemplatesGet(w http.ResponseWriter, r *http.Request, role string) {
	ids, err := s.Model.Ignition.GetTemplateIDs(r.Context(), role)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	renderJSON(w, ids, http.StatusOK)
}

func (s Server) handleIgnitionTemplatesPut(w http.ResponseWriter, r *http.Request, role string) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	id, err := s.Model.Ignition.PutTemplate(r.Context(), role, string(body))
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

func (s Server) serveIgnition(w http.ResponseWriter, r *http.Request, role, id, serial string) {
	tmpl, err := s.Model.Ignition.GetTemplate(r.Context(), role, id)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	q := sabakan.QueryBySerial(serial)
	ms, err := s.Model.Machine.Query(r.Context(), q)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	if len(ms) == 0 {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	ign, err := renderIgnition(tmpl, ms[0])
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte(ign))
	if err != nil {
		fields := cmd.FieldsFromContext(r.Context())
		fields[log.FnError] = err.Error()
		log.Error("failed to write response for GET /boot/ignitions", fields)
	}
}

func renderIgnition(tmpl string, m *sabakan.Machine) (string, error) {
	t, err := template.New("ignition").Parse(tmpl)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, m)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
