package web

// intern yattakoto
// server.go
//   add switch case p == "version"

// web/version.go
// add method to Server struct

import (
	"net/http"

	"github.com/cybozu-go/sabakan"
)

func (s Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":

		version := map[string]string{"version": sabakan.Version}

		renderJSON(w, version, http.StatusOK)
	}
}
