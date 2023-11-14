package web

import (
	"io"
	"net/http"
	"strings"

	"github.com/cybozu-go/sabakan/v3"
)

func (s Server) handleLabels(w http.ResponseWriter, r *http.Request) {
	args := strings.SplitN(r.URL.Path[len("/api/v1/labels/"):], "/", 2)
	if len(args) != 2 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}
	if !sabakan.IsValidLabelName(args[1]) {
		renderError(r.Context(), w, BadRequest("invalid label name"))
		return
	}

	switch r.Method {
	case "PUT":
		s.handleLabelsPut(w, r, args[0], args[1])
		return
	case "DELETE":
		s.handleLabelsDelete(w, r, args[0], args[1])
		return
	}

	renderError(r.Context(), w, APIErrBadMethod)
}

func (s Server) handleLabelsPut(w http.ResponseWriter, r *http.Request, serial, label string) {
	value, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	if !sabakan.IsValidLabelValue(string(value)) {
		renderError(r.Context(), w, BadRequest("invalid label value"))
		return
	}

	err = s.Model.Machine.PutLabel(r.Context(), serial, label, string(value))
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
