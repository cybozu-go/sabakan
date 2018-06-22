package web

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cybozu-go/sabakan"
)

const (
	// HeaderSabactlUser is the HTTP header name to tell which user run sabactl.
	HeaderSabactlUser = "X-Sabakan-User"
)

var (
	hostnameAtStartup string
)

func init() {
	hostnameAtStartup, _ = os.Hostname()
}

// Server is the sabakan server.
type Server struct {
	Model          sabakan.Model
	MyURL          *url.URL
	IPXEFirmware   string
	AllowedRemotes []*net.IPNet
}

// Handler implements http.Handler
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/v1/") {
		s.handleAPIV1(w, r)
		return
	}

	renderError(r.Context(), w, APIErrNotFound)
}

func auditContext(r *http.Request) context.Context {
	ctx := r.Context()

	u := r.Header.Get(HeaderSabactlUser)
	if len(u) > 0 {
		ctx = context.WithValue(ctx, sabakan.AuditKeyUser, u)
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if len(ip) > 0 {
		ctx = context.WithValue(ctx, sabakan.AuditKeyIP, ip)
	}

	if len(hostnameAtStartup) > 0 {
		ctx = context.WithValue(ctx, sabakan.AuditKeyHost, hostnameAtStartup)
	} else {
		ctx = context.WithValue(ctx, sabakan.AuditKeyHost, r.Host)
	}

	return ctx
}

func (s Server) handleAPIV1(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/api/v1/"):]

	if !s.hasPermission(r) {
		renderError(r.Context(), w, APIErrForbidden)
		return
	}

	r = r.WithContext(auditContext(r))

	switch {
	case p == "assets" || strings.HasPrefix(p, "assets/"):
		s.handleAssets(w, r)
	case p == "boot/ipxe.efi":
		http.ServeFile(w, r, s.IPXEFirmware)
	case strings.HasPrefix(p, "boot/coreos/"):
		s.handleCoreOS(w, r)
	case strings.HasPrefix(p, "boot/ignitions/"):
		s.handleIgnitions(w, r)
	case p == "config/dhcp":
		s.handleConfigDHCP(w, r)
	case p == "config/ipam":
		s.handleConfigIPAM(w, r)
	case strings.HasPrefix(p, "crypts/"):
		s.handleCrypts(w, r)
	case strings.HasPrefix(p, "ignitions/"):
		s.handleIgnitionTemplates(w, r)
	case p == "images/coreos" || strings.HasPrefix(p, "images/coreos/"):
		s.handleImages(w, r)
	case strings.HasPrefix(p, "machines"):
		s.handleMachines(w, r)
	case strings.HasPrefix(p, "state/"):
		s.handleState(w, r)
	default:
		renderError(r.Context(), w, APIErrNotFound)
	}

}

// hasPermission returns true if the request has a permission to the resource
func (s Server) hasPermission(r *http.Request) bool {
	p := r.URL.Path[len("/api/v1/"):]
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	if strings.HasPrefix(p, "crypts/") && r.Method != http.MethodDelete {
		return true
	}
	rhost, _, err := net.SplitHostPort(r.RemoteAddr)
	if rhost == "" || err != nil {
		return false
	}
	for _, allowed := range s.AllowedRemotes {
		if allowed.Contains(net.ParseIP(rhost)) {
			return true
		}
	}
	return false
}
