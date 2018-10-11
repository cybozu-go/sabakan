package web

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/cybozu-go/sabakan"
)

func (s Server) handleState(w http.ResponseWriter, r *http.Request) {
	serial := r.URL.Path[len("/api/v1/state/"):]
	if len(serial) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.handleStateGet(w, r, serial)
		return
	case "PUT":
		s.handleStatePut(w, r, serial)
		return
	}

	renderError(r.Context(), w, APIErrBadMethod)
}

func (s Server) handleStateGet(w http.ResponseWriter, r *http.Request, serial string) {
	m, err := s.Model.Machine.Get(r.Context(), serial)
	switch err {
	case sabakan.ErrNotFound:
		renderError(r.Context(), w, APIErrNotFound)
		return
	case nil:
	default:
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	w.Header().Set("content-type", "text/plain")
	io.WriteString(w, m.Status.State.String())
}

func (s Server) handleStatePut(w http.ResponseWriter, r *http.Request, serial string) {
	state, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 128))
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	ms := sabakan.MachineState(state)
	switch ms {
	case sabakan.StateUninitialized, sabakan.StateHealthy, sabakan.StateUnhealthy, sabakan.StateUnreachable, sabakan.StateUpdating, sabakan.StateRetiring:
	default:
		renderError(r.Context(), w, BadRequest("invalid state: "+string(state)))
		return
	}
	err = s.Model.Machine.SetState(r.Context(), serial, ms)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}
