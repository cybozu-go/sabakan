package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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
	// see https://coreos.com/ignition/docs/latest/
	ign := `
{
  "ignition": { "version": "2.2.0" },
  "storage": {
    "files": [{
      "filesystem": "root",
      "path": "/etc/hostname",
      "mode": 420,
      "contents": { "source": %s }
    }]
  },
  "passwd": {
    "users": [
      {
        "groups": [ "sudo" ],
        "name": "cybozu",
        "passwordHash": "$6$rounds=4096$m3AVOWeB$EPystoHozf.eJNCm4tWyRHpJzgTDymYuGOONWxRN8uk4amLvxwB4Pc7.tEkZdeXewoVEBEX5ujUon9wSpEf1N."
       }
    ]
  }
}
`
	hostname, err := json.Marshal("data:,core-" + strings.ToLower(serial))
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(ign, hostname)))
}
