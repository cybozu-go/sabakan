package web

import (
	"encoding/json"
	"net/http"

	"github.com/cybozu-go/sabakan"
)

func (s Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handleConfigGet(w, r)
		return
	case "PUT":
		s.handleConfigPut(w, r)
		return
	}

	renderError(r.Context(), w, APIErrBadMethod)
}

func (s Server) handleConfigGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	config, err := s.Model.Config.GetConfig(ctx)
	if err != nil {
		renderError(ctx, w, InternalServerError(err))
		return
	}
	if config == nil {
		renderError(ctx, w, APIErrNotFound)
		return
	}

	renderJSON(w, config, http.StatusOK)
}

func (s Server) handleConfigPut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var sc sabakan.IPAMConfig
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

	err = s.Model.Config.PutConfig(ctx, &sc)
	if err != nil {
		renderError(ctx, w, InternalServerError(err))
		return
	}
	renderJSON(w, nil, http.StatusOK)
}
