package web

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/cybozu-go/sabakan/v3"
)

const (
	// iPXE script specs can be found at http://ipxe.org/cfg and http://ipxe.org/cmd
	redirectiPXETemplate = `#!ipxe
chain %s/${serial}
`

	coreOSiPXETemplate = `#!ipxe

set base-url %s
set ignition-id %s
kernel ${base-url}/coreos/kernel initrd=initrd.gz coreos.first_boot=1 coreos.config.url=${base-url}/ignitions/${serial}/${ignition-id} %s
initrd ${base-url}/coreos/initrd.gz
boot
`
)

func (s Server) handleCoreOS(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/boot/coreos/"):], "/")

	if r.Method != "GET" && r.Method != "HEAD" {
		renderError(r.Context(), w, APIErrBadMethod)
		return
	}

	switch params[0] {
	case "ipxe":
		if len(params) == 2 {
			s.handleCoreOSiPXEWithSerial(w, r, params[1])
		} else {
			s.handleCoreOSiPXE(w, r)
		}
	case "kernel":
		s.handleCoreOSKernel(w, r)
	case "initrd.gz":
		s.handleCoreOSInitRD(w, r)
	default:
		renderError(r.Context(), w, APIErrNotFound)
	}
}

func (s Server) handleCoreOSiPXE(w http.ResponseWriter, r *http.Request) {
	u := *s.MyURL
	u.Path = path.Join("/api/v1/boot/coreos/ipxe")
	ipxe := fmt.Sprintf(redirectiPXETemplate, u.String())

	w.Header().Set("Content-Type", "text/plain; charset=ASCII")
	w.Write([]byte(ipxe))
}

func (s Server) handleCoreOSiPXEWithSerial(w http.ResponseWriter, r *http.Request, serial string) {
	m, err := s.Model.Machine.Get(r.Context(), serial)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	role := m.Spec.Role
	ids, err := s.Model.Ignition.GetTemplateIDs(r.Context(), role)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	if len(ids) == 0 {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}

	params, err := s.Model.KernelParams.GetParams(r.Context(), "coreos")
	if err != sabakan.ErrNotFound && err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	u := *s.MyURL
	u.Path = path.Join("/api/v1/boot")
	ipxe := fmt.Sprintf(coreOSiPXETemplate, u.String(), ids[len(ids)-1], params)

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
