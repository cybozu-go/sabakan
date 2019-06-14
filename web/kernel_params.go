package web

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/v2"
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

	renderError(r.Context(), w, APIErrBadMethod)
}

func (s Server) handleKernelParamsGet(w http.ResponseWriter, r *http.Request, os string) {
	ctx := r.Context()
	kernelParams, err := s.Model.KernelParams.GetParams(ctx, os)
	if err == sabakan.ErrNotFound {
		renderError(ctx, w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(ctx, w, InternalServerError(err))
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

	kp, err := ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 4096))
	if err != nil {
		renderError(ctx, w, BadRequest(err.Error()))
		return
	}
	if !sabakan.IsValidKernelParams(string(kp)) {
		renderError(ctx, w, BadRequest("kernel params is not valid"))
		return
	}

	err = s.Model.KernelParams.PutParams(ctx, os, string(kp))
	if err != nil {
		renderError(ctx, w, InternalServerError(err))
		return
	}

	renderJSON(w, nil, http.StatusOK)
}
