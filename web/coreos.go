package web

import (
	"fmt"
	"net/http"
	"path/filepath"
)

const (
	coreOSImageDir = "/var/www/assets/coreos/1576.5.0"

	coreOSiPXETemplate = `#!ipxe

set base-url %s
kernel ${base-url}/coreos/kernel initrd=initrd.gz coreos.first_boot=1 coreos.config.url=${base-url}/ignitions/${serial}
initrd ${base-url}/coreos/initrd.gz
boot
`
)

func (s Server) handleCoreOS(w http.ResponseWriter, r *http.Request) {
	item := r.URL.Path[len("/api/v1/boot/coreos/"):]

	if r.Method != "GET" {
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
	}

	renderError(r.Context(), w, APIErrNotFound)
}

func (s Server) handleCoreOSiPXE(w http.ResponseWriter, r *http.Request) {
	// iPXE script specs can be found at http://ipxe.org/cfg
	baseURL := fmt.Sprintf("http://%s/api/v1/boot", r.Host)
	ipxe := fmt.Sprintf(coreOSiPXETemplate, baseURL)

	w.Header().Set("Content-Type", "text/plain; charset=ASCII")
	w.Write([]byte(ipxe))
}

func (s Server) handleCoreOSKernel(w http.ResponseWriter, r *http.Request) {
	p := filepath.Join(coreOSImageDir, "coreos_production_pxe.vmlinuz")
	http.ServeFile(w, r, p)
}

func (s Server) handleCoreOSInitRD(w http.ResponseWriter, r *http.Request) {
	p := filepath.Join(coreOSImageDir, "coreos_production_pxe_image.cpio.gz")
	http.ServeFile(w, r, p)
}
