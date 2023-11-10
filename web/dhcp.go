package web

import (
	"encoding/json"
	"net/http"

	"github.com/cybozu-go/sabakan/v3"
)

func (s Server) handleConfigDHCP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleConfigDHCPGet(w, r)
		return
	case "PUT":
		s.handleConfigDHCPPut(w, r)
		return
	}

	renderError(r.Context(), w, APIErrBadMethod)
}

func (s Server) handleConfigDHCPGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config, err := s.Model.DHCP.GetConfig()
	if err != nil {
		renderError(ctx, w, APIErrNotFound)
		return
	}

	renderJSON(w, config, http.StatusOK)
}

func (s Server) handleConfigDHCPPut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sc sabakan.DHCPConfig
	err := json.NewDecoder(r.Body).Decode(&sc)
	if err != nil {
		renderError(ctx, w, APIErrBadRequest)
		return
	}
	err = sc.Validate()
	if err != nil {
		renderError(ctx, w, BadRequest(err.Error()))
		return
	}

	err = s.Model.DHCP.PutConfig(ctx, &sc)
	if err != nil {
		renderError(ctx, w, InternalServerError(err))
		return
	}
	renderJSON(w, nil, http.StatusOK)
}
