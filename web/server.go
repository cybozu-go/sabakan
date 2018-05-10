package web

import (
	"net/http"
	"strings"

	"github.com/cybozu-go/sabakan"
)

// Server is the sabakan server.
type Server struct {
	Model sabakan.Model
}

// Handler implements http.Handler
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/v1/") {
		s.handleAPIV1(w, r)
		return
	}

	renderError(r.Context(), w, APIErrNotFound)
}

func (s Server) handleAPIV1(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/api/v1/"):]

	switch {
	case p == "config":
		s.handleConfig(w, r)
		return
	case strings.HasPrefix(p, "crypts"):
		s.handleCrypts(w, r)
		return
	case strings.HasPrefix(p, "ignitions"):
		//s.handleIgnitions(w, r)
		//return
	case strings.HasPrefix(p, "machines"):
		s.handleMachines(w, r)
		return
	}

	renderError(r.Context(), w, APIErrNotFound)
}
