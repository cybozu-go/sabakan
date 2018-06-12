package web

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cybozu-go/sabakan"
)

const (
	maxAssetSize = 2 << 30
)

func (s Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/assets" {
		s.handleAssetsIndex(w, r)
		return
	}

	params := strings.Split(r.URL.Path[len("/api/v1/assets/"):], "/")
	name := params[0]

	switch r.Method {
	case "GET", "HEAD":
		switch len(params) {
		case 1:
			s.handleAssetsGet(w, r, name)
			return
		case 2:
			if params[1] == "meta" {
				s.handleAssetsInfo(w, r, name)
				return
			}
		}
		renderError(r.Context(), w, APIErrBadRequest)
	case "PUT":
		s.handleAssetsPut(w, r, name)
	case "DELETE":
		s.handleAssetsDelete(w, r, name)
	default:
		renderError(r.Context(), w, APIErrBadMethod)
	}
}

func (s Server) handleAssetsIndex(w http.ResponseWriter, r *http.Request) {
	index, err := s.Model.Asset.GetIndex(r.Context())
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	renderJSON(w, index, http.StatusOK)
}

func (s Server) handleAssetsGet(w http.ResponseWriter, r *http.Request, name string) {
	f := func(modtime time.Time, contentType string, content io.ReadSeeker) {
		w.Header().Set("content-type", contentType)
		http.ServeContent(w, r, name, modtime, content)
	}
	err := s.Model.Asset.Get(r.Context(), name, f)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}

func (s Server) handleAssetsInfo(w http.ResponseWriter, r *http.Request, name string) {
	asset, err := s.Model.Asset.GetInfo(r.Context(), name)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	renderJSON(w, asset, http.StatusOK)
}

func (s Server) handleAssetsPut(w http.ResponseWriter, r *http.Request, name string) {
	contentType := r.Header.Get("content-type")
	if len(contentType) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}
	if r.ContentLength < 0 {
		renderError(r.Context(), w, APIErrLengthRequired)
		return
	}
	if r.ContentLength > maxAssetSize {
		renderError(r.Context(), w, APIErrTooLargeAsset)
		return
	}
	status, err := s.Model.Asset.Put(r.Context(), name, contentType, r.Body)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	renderJSON(w, status, status.Status)
}

func (s Server) handleAssetsDelete(w http.ResponseWriter, r *http.Request, name string) {
	err := s.Model.Asset.Delete(r.Context(), name)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}
