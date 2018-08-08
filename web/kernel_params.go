package web

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

func (s Server) handleKernelParams(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path[len("/api/v1/kernel_params/"):], "/")

	if len(params) < 1 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	os := params[0]
	if !sabakan.IsValidImageOS(os) {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		s.handleKernelParamsGet(w, r, os)
		return
	case "PUT":
		s.handleKernelParamsPut(w, r, os)
		return
	}
}

func (s Server) handleKernelParamsGet(w http.ResponseWriter, r *http.Request, os string) {
	ctx := r.Context()
	kernelParams, err := s.Model.KernelParams.GetParams(ctx, os)
	if err != nil {
		renderError(ctx, w, APIErrNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	_, err = w.Write([]byte(kernelParams))
	if err != nil {
		log.Error("failed to output text", map[string]interface{}{
			log.FnError: err.Error(),
		})
	}
}

func (s Server) handleKernelParamsPut(w http.ResponseWriter, r *http.Request, os string) {
	ctx := r.Context()

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		renderError(ctx, w, BadRequest(err.Error()))
		return
	}
	if len(data) == 0 {
		renderError(ctx, w, BadRequest("Kernel parameters is empty."))
		return
	}

	err = s.Model.KernelParams.PutParams(ctx, os, sabakan.KernelParams(data))
	if err != nil {
		renderError(ctx, w, InternalServerError(err))
		return
	}

	renderJSON(w, nil, http.StatusOK)
}
