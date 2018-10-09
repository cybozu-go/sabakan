package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cybozu-go/sabakan"
)

func (s Server) handleLabels(w http.ResponseWriter, r *http.Request) {
	args := strings.SplitN(r.URL.Path[len("/api/v1/labels/"):], "/", 2)
	if len(args) == 0 || len(args[0]) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	switch r.Method {
	case "PUT":
		s.handleLabelsPut(w, r, args[0])
		return
	case "DELETE":
		if len(args) != 2 {
			renderError(r.Context(), w, APIErrBadRequest)
			return
		}
		s.handleLabelsDelete(w, r, args[0], args[1])
		return
	}
}

func (s Server) handleLabelsPut(w http.ResponseWriter, r *http.Request, serial string) {
	var labels map[string]string
	err := json.NewDecoder(r.Body).Decode(&labels)
	if err != nil {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	err = s.Model.Machine.AddLabels(r.Context(), serial, labels)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}

func (s Server) handleLabelsDelete(w http.ResponseWriter, r *http.Request, serial, label string) {
	err := s.Model.Machine.DeleteLabel(r.Context(), serial, label)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}
