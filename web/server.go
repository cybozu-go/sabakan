package web

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/cybozu-go/sabakan"
)

// Server is the sabakan server.
type Server struct {
	Model         sabakan.Model
	MyURL         *url.URL
	IPXEFirmware  string
	AllowdRemotes []*net.IPNet
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

	if !s.hasPermission(r) {
		renderError(r.Context(), w, APIErrForbidden)
	}

	switch {
	case p == "config/dhcp":
		s.handleConfigDHCP(w, r)
		return
	case p == "config/ipam":
		s.handleConfigIPAM(w, r)
		return
	case strings.HasPrefix(p, "crypts/"):
		s.handleCrypts(w, r)
		return
	case p == "boot/ipxe.efi":
		http.ServeFile(w, r, s.IPXEFirmware)
		return
	case strings.HasPrefix(p, "boot/coreos/"):
		s.handleCoreOS(w, r)
		return
	case strings.HasPrefix(p, "boot/ignitions/"):
		s.handleIgnitions(w, r)
		return
	case strings.HasPrefix(p, "ignitions/"):
		s.handleIgnitionTemplates(w, r)
		return
	case strings.HasPrefix(p, "machines"):
		s.handleMachines(w, r)
		return
	case p == "images/coreos" || strings.HasPrefix(p, "images/coreos/"):
		s.handleImages(w, r)
		return
	}

	renderError(r.Context(), w, APIErrNotFound)
}

// hasPermission returns true if the request has a permission to the resource
func (s Server) hasPermission(r *http.Request) bool {
	p := r.URL.Path[len("/api/v1/"):]
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	if strings.HasPrefix(p, "boot/") || p == "crypts" {
		return true
	}
	rhost := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	if rhost == nil {
		return false
	}
	for _, allowed := range s.AllowdRemotes {
		if allowed.Contains(rhost) {
			return true
		}
	}
	return false
}
