package web

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cybozu-go/sabakan"
)

func (s Server) handleRetireDate(w http.ResponseWriter, r *http.Request) {
	serial := r.URL.Path[len("/api/v1/retire-date/"):]
	if len(serial) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	if r.Method != http.MethodPut {
		renderError(r.Context(), w, APIErrBadMethod)
		return
	}

	data, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
	if err != nil {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}
	date, err := time.Parse(time.RFC3339Nano, string(bytes.TrimSpace(data)))
	if err != nil {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	err = s.Model.Machine.SetRetireDate(r.Context(), serial, date)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}
