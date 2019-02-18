package web

import (
	"net/http"
	"strings"

	"github.com/cybozu-go/sabakan/v2"
)

func (s Server) handleImages(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/images/"):], "/")

	if len(params) < 1 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	os := params[0]
	if !sabakan.IsValidImageOS(os) {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	if len(params) == 1 && r.Method == "GET" {
		s.handleImageIndexGet(w, r, os)
		return
	}

	if len(params) != 2 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}
	id := params[1]
	if !sabakan.IsValidImageID(id) {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.handleImagesGet(w, r, os, id)
		return
	case "PUT":
		s.handleImagesPut(w, r, os, id)
		return
	case "DELETE":
		s.handleImagesDelete(w, r, os, id)
		return
	}

	renderError(r.Context(), w, APIErrBadMethod)
}

func (s Server) handleImageIndexGet(w http.ResponseWriter, r *http.Request, os string) {
	index, err := s.Model.Image.GetIndex(r.Context(), os)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	renderJSON(w, index, http.StatusOK)
}

func (s Server) handleImagesGet(w http.ResponseWriter, r *http.Request, os, id string) {
	w.Header().Set("content-type", "application/tar")
	err := s.Model.Image.Download(r.Context(), os, id, w)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}

func (s Server) handleImagesPut(w http.ResponseWriter, r *http.Request, os, id string) {
	err := s.Model.Image.Upload(r.Context(), os, id, r.Body)
	switch err {
	case sabakan.ErrConflicted:
		renderError(r.Context(), w, APIErrConflict)
	case sabakan.ErrBadRequest:
		renderError(r.Context(), w, APIErrBadRequest)
	case nil:
		w.WriteHeader(http.StatusCreated)
	default:
		renderError(r.Context(), w, InternalServerError(err))
	}
}

func (s Server) handleImagesDelete(w http.ResponseWriter, r *http.Request, os, id string) {
	err := s.Model.Image.Delete(r.Context(), os, id)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}
