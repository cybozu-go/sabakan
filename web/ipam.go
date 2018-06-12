package web

import (
	"encoding/json"
	"net/http"

	"github.com/cybozu-go/sabakan"
)

func (s Server) handleConfigIPAM(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleConfigIPAMGet(w, r)
		return
	case "PUT":
		s.handleConfigIPAMPut(w, r)
		return
	}

	renderError(r.Context(), w, APIErrBadMethod)
}

func (s Server) handleConfigIPAMGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config, err := s.Model.IPAM.GetConfig()
	if err != nil {
		renderError(ctx, w, APIErrNotFound)
		return
	}

	renderJSON(w, config, http.StatusOK)
}

func (s Server) handleConfigIPAMPut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sc sabakan.IPAMConfig
	err := json.NewDecoder(r.Body).Decode(&sc)
	if err != nil {
		renderError(ctx, w, BadRequest(err.Error()))
		return
	}
	err = sc.Validate()
	if err != nil {
		renderError(ctx, w, BadRequest(err.Error()))
		return
	}

	err = s.Model.IPAM.PutConfig(ctx, &sc)
	if err != nil {
		renderError(ctx, w, InternalServerError(err))
		return
	}
	renderJSON(w, nil, http.StatusOK)
}
