package web

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/v3"
	"github.com/cybozu-go/well"
)

func (s Server) handleCrypts(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/crypts/"):], "/")

	if len(params) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.handleCryptsGet(w, r, params)
		return
	case "PUT":
		s.handleCryptsPut(w, r, params)
		return
	case "DELETE":
		s.handleCryptsDelete(w, r, params[0])
		return
	}

	renderError(r.Context(), w, APIErrBadMethod)
}

func (s Server) handleCryptsGet(w http.ResponseWriter, r *http.Request, params []string) {
	if len(params) != 2 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	serial := params[0]
	p := params[1]

	key, err := s.Model.Storage.GetEncryptionKey(r.Context(), serial, p)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	if key == nil {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(key)))
	_, err = w.Write(key)
	if err != nil {
		fields := well.FieldsFromContext(r.Context())
		fields[log.FnError] = err.Error()
		log.Error("failed to write response for GET /crypts", fields)
	}
}

func (s Server) handleCryptsPut(w http.ResponseWriter, r *http.Request, params []string) {
	if len(params) != 2 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	serial := params[0]
	p := params[1]

	keyData, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4096))
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	if len(keyData) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	err = s.Model.Storage.PutEncryptionKey(r.Context(), serial, p, keyData)
	switch err {
	case sabakan.ErrConflicted:
		renderError(r.Context(), w, APIErrConflict)
		return
	case nil:
		// do nothing
	default:
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	resp := make(map[string]interface{})
	resp["status"] = http.StatusCreated
	resp["path"] = p

	renderJSON(w, resp, http.StatusCreated)
}

func (s Server) handleCryptsDelete(w http.ResponseWriter, r *http.Request, serial string) {
	keys, err := s.Model.Storage.DeleteEncryptionKeys(r.Context(), serial)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	} else if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	renderJSON(w, keys, http.StatusOK)
}

func (s Server) handleCryptSetup(w http.ResponseWriter, r *http.Request) {
	if s.CryptSetup == "" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, s.CryptSetup)
}
