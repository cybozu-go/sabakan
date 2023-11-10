package web

import (
	"net/http"

	"github.com/cybozu-go/sabakan/v3"
)

func (s Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		version := map[string]string{"version": sabakan.Version}
		renderJSON(w, version, http.StatusOK)
		return
	}

	renderError(r.Context(), w, APIErrBadMethod)
}
