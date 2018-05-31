package web

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/cybozu-go/sabakan"
)

const (
	// iPXE script specs can be found at http://ipxe.org/cfg
	coreOSiPXETemplate = `#!ipxe

set base-url %s
kernel ${base-url}/coreos/kernel initrd=initrd.gz coreos.first_boot=1 coreos.config.url=${base-url}/ignitions/${serial} %s
initrd ${base-url}/coreos/initrd.gz
boot
`
)

func (s Server) handleCoreOS(w http.ResponseWriter, r *http.Request) {
	item := r.URL.Path[len("/api/v1/boot/coreos/"):]

	if r.Method != "GET" && r.Method != "HEAD" {
		renderError(r.Context(), w, APIErrBadMethod)
		return
	}

	switch item {
	case "ipxe":
		s.handleCoreOSiPXE(w, r)
	case "kernel":
		s.handleCoreOSKernel(w, r)
	case "initrd.gz":
		s.handleCoreOSInitRD(w, r)
	default:
		renderError(r.Context(), w, APIErrNotFound)
	}
}

func (s Server) handleCoreOSiPXE(w http.ResponseWriter, r *http.Request) {
	console := ""
	if r.URL.Query().Get("serial") == "1" {
		console = "console=ttyS0"
	}

	u := *s.MyURL
	u.Path = path.Join("/api/v1/boot")
	ipxe := fmt.Sprintf(coreOSiPXETemplate, u.String(), console)

	w.Header().Set("Content-Type", "text/plain; charset=ASCII")
	w.Write([]byte(ipxe))
}

func (s Server) handleCoreOSKernel(w http.ResponseWriter, r *http.Request) {
	f := func(modtime time.Time, content io.ReadSeeker) {
		http.ServeContent(w, r, sabakan.ImageKernelFilename, modtime, content)
	}
	w.Header().Set("content-type", "application/octet-stream")
	err := s.Model.Image.ServeFile(r.Context(), "coreos", sabakan.ImageKernelFilename, f)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}

func (s Server) handleCoreOSInitRD(w http.ResponseWriter, r *http.Request) {
	f := func(modtime time.Time, content io.ReadSeeker) {
		http.ServeContent(w, r, sabakan.ImageInitrdFilename, modtime, content)
	}
	w.Header().Set("content-type", "application/octet-stream")
	err := s.Model.Image.ServeFile(r.Context(), "coreos", sabakan.ImageInitrdFilename, f)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}
