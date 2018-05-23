package web

import "net/http"

func (s Server) handleIgnitions(w http.ResponseWriter, r *http.Request) {
	serial := r.URL.Path[len("/api/v1/ignitions/"):]
	if len(serial) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	if r.Method != "GET" {
		renderError(r.Context(), w, APIErrBadMethod)
		return
	}

	s.serveIgnition(w, r, serial)
}

func (s Server) serveIgnition(w http.ResponseWriter, r *http.Request, serial string) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`
{
  "ignition": { "version": "2.2.0" },
  "storage": {
    "files": [{
      "filesystem": "root",
      "path": "/etc/hostname",
      "mode": 420,
      "contents": { "source": "data:,core1" }
    }]
  }
}
`))
}
