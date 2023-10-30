package web

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/cybozu-go/sabakan/v2"
	"github.com/cybozu-go/sabakan/v2/gql/graph"
	"github.com/cybozu-go/sabakan/v2/gql/graph/generated"
	"github.com/cybozu-go/sabakan/v2/metrics"
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

// recorderWriter extends http.ReponseWrite to log the status code in WriteHeader().
// The status code is used later for logging.
type recorderWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *recorderWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Server is the sabakan server.
type Server struct {
	Model          sabakan.Model
	MyURL          *url.URL
	MyURLHTTPS     *url.URL
	IPXEFirmware   string
	CryptSetup     string
	AllowedRemotes []*net.IPNet
	Counter        *metrics.APICounter

	graphQL    http.Handler
	playground http.HandlerFunc

	TLSServer bool
}

// NewServer constructs Server instance
func NewServer(model sabakan.Model, ipxePath, cryptsetupPath string,
	advertiseURL, advertiseURLHTTPS *url.URL, allowedIPs []*net.IPNet, enablePlayground bool, counter *metrics.APICounter, tlsServer bool) *Server {
	graphQL := handler.NewDefaultServer(generated.NewExecutableSchema(
		generated.Config{
			Resolvers: &graph.Resolver{Model: model},
		},
	))
	s := &Server{
		Model:          model,
		IPXEFirmware:   ipxePath,
		CryptSetup:     cryptsetupPath,
		MyURL:          advertiseURL,
		MyURLHTTPS:     advertiseURLHTTPS,
		AllowedRemotes: allowedIPs,
		Counter:        counter,
		graphQL:        graphQL,
		TLSServer:      tlsServer,
	}

	if enablePlayground {
		s.playground = playground.Handler("GraphQL playground", "/graphql")
	}
	return s
}

// ServeHTTP implements http.Handler.ServeHTTP
func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w2 := &recorderWriter{ResponseWriter: w}
	if s.TLSServer {
		s.serveHTTPS(w2, r)
	} else {
		s.serveHTTP(w2, r)
	}
	if s.Counter != nil {
		s.Counter.Inc(w2.statusCode, r.URL.Path, r.Method)
	}
}

func (s Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/v1/") {
		s.handleAPIV1(w, r)
		return
	}

	if r.URL.Path == "/graphql" {
		s.graphQL.ServeHTTP(w, r)
		return
	}

	if r.URL.Path == "/playground" && s.playground != nil {
		s.playground.ServeHTTP(w, r)
		return
	}

	if r.URL.Path == "/version" {
		s.handleVersion(w, r)
		return
	}

	if r.URL.Path == "/health" {
		s.handleHealth(w, r)
		return
	}

	renderError(r.Context(), w, APIErrNotFound)
}

func (s Server) serveHTTPS(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/v1/") {
		s.handleAPIV1HTTPS(w, r)
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
	case p == "cryptsetup":
		s.handleCryptSetup(w, r)
	case strings.HasPrefix(p, "ignitions/"):
		s.handleIgnitionTemplates(w, r)
	case p == "images/coreos" || strings.HasPrefix(p, "images/coreos/"):
		s.handleImages(w, r)
	case p == "logs":
		s.handleLogs(w, r)
	case strings.HasPrefix(p, "machines"):
		s.handleMachines(w, r)
	case strings.HasPrefix(p, "state/"):
		s.handleState(w, r)
	case strings.HasPrefix(p, "labels/"):
		s.handleLabels(w, r)
	case strings.HasPrefix(p, "retire-date/"):
		s.handleRetireDate(w, r)
	case strings.HasPrefix(p, "kernel_params/"):
		s.handleKernelParams(w, r)
	default:
		renderError(r.Context(), w, APIErrNotFound)
	}
}

func (s Server) handleAPIV1HTTPS(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path[len("/api/v1/"):]

	if !s.hasPermission(r) {
		renderError(r.Context(), w, APIErrForbidden)
		return
	}

	r = r.WithContext(auditContext(r))

	switch {
	case strings.HasPrefix(p, "crypts/"):
		s.handleCrypts(w, r)
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
