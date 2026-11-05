package web

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/cybozu-go/sabakan/v3"
)

const (
	maxAssetSize = 2 << 31

	// assetOptionAllowedRoles restricts GET and HEAD of the asset to machines whose role is in the comma-separated list.
	assetOptionAllowedRoles = "allowed-roles"
)

func (s Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/assets" {
		s.handleAssetsIndex(w, r)
		return
	}

	params := strings.Split(r.URL.Path[len("/api/v1/assets/"):], "/")
	name := params[0]

	switch r.Method {
	case "GET", "HEAD":
		switch len(params) {
		case 1:
			s.handleAssetsGet(w, r, name)
			return
		case 2:
			if params[1] == "meta" {
				s.handleAssetsInfo(w, r, name)
				return
			}
		}
		renderError(r.Context(), w, APIErrBadRequest)
	case "PUT":
		s.handleAssetsPut(w, r, name)
	case "DELETE":
		s.handleAssetsDelete(w, r, name)
	default:
		renderError(r.Context(), w, APIErrBadMethod)
	}
}

func (s Server) handleAssetsIndex(w http.ResponseWriter, r *http.Request) {
	index, err := s.Model.Asset.GetIndex(r.Context())
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	renderJSON(w, index, http.StatusOK)
}

type assetHandler struct {
	s Server
	w http.ResponseWriter
	r *http.Request
}

func (h assetHandler) ServeContent(asset *sabakan.Asset, content io.ReadSeeker) {
	allowed, err := h.s.isAllowedAssetClient(h.r.Context(), asset, h.r.RemoteAddr)
	if err != nil {
		renderError(h.r.Context(), h.w, InternalServerError(err))
		return
	}
	if !allowed {
		renderError(h.r.Context(), h.w, APIErrForbidden)
		return
	}

	header := h.w.Header()
	header.Set("content-type", asset.ContentType)
	header.Set("X-Sabakan-Asset-ID", strconv.Itoa(asset.ID))
	header.Set("X-Sabakan-Asset-SHA256", asset.Sha256)
	http.ServeContent(h.w, h.r, asset.Name, asset.Date, content)
}

func (h assetHandler) Redirect(u string) {
	http.Redirect(h.w, h.r, u, http.StatusFound)
}

func (s Server) handleAssetsGet(w http.ResponseWriter, r *http.Request, name string) {
	err := s.Model.Asset.Get(r.Context(), name, assetHandler{s, w, r})
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}

func (s Server) handleAssetsInfo(w http.ResponseWriter, r *http.Request, name string) {
	asset, err := s.Model.Asset.GetInfo(r.Context(), name)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}

	renderJSON(w, asset, http.StatusOK)
}

func (s Server) handleAssetsPut(w http.ResponseWriter, r *http.Request, name string) {
	contentType := r.Header.Get("content-type")
	if len(contentType) == 0 {
		renderError(r.Context(), w, APIErrBadRequest)
		return
	}
	if r.ContentLength < 0 {
		renderError(r.Context(), w, APIErrLengthRequired)
		return
	}
	if r.ContentLength > maxAssetSize {
		renderError(r.Context(), w, APIErrTooLargeAsset)
		return
	}

	sum := r.Header.Get("X-Sabakan-Asset-SHA256")
	var csum []byte
	if len(sum) > 0 {
		c, err := hex.DecodeString(sum)
		if err != nil {
			renderError(r.Context(), w, BadRequest("bad checksum: "+err.Error()))
			return
		}
		csum = c
	}

	options := make(map[string]string)
	for k, v := range r.Header {
		optionHeaderPrefix := "x-sabakan-asset-options-"
		if strings.HasPrefix(strings.ToLower(k), optionHeaderPrefix) {
			key := strings.ToLower(k[len(optionHeaderPrefix):])
			if len(key) == 0 {
				continue
			}
			if !sabakan.IsValidLabelName(key) {
				renderError(r.Context(), w, BadRequest("invalid option key"+key))
				return
			}
			if !isValidAssetOptionValue(key, v[0]) {
				renderError(r.Context(), w, BadRequest("invalid option value"+v[0]))
				return
			}
			options[key] = v[0]
		}
	}

	status, err := s.Model.Asset.Put(r.Context(), name, contentType, csum, options, r.Body)
	if err == sabakan.ErrConflicted {
		renderError(r.Context(), w, APIErrConflict)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
		return
	}
	renderJSON(w, status, status.Status)
}

func (s Server) handleAssetsDelete(w http.ResponseWriter, r *http.Request, name string) {
	err := s.Model.Asset.Delete(r.Context(), name)
	if err == sabakan.ErrNotFound {
		renderError(r.Context(), w, APIErrNotFound)
		return
	}
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}

func isValidAssetOptionValue(key, value string) bool {
	if key != assetOptionAllowedRoles {
		return sabakan.IsValidLabelValue(value)
	}
	for role := range strings.SplitSeq(value, ",") {
		if !sabakan.IsValidRole(role) {
			return false
		}
	}
	return true
}

func (s Server) isAllowedAssetClient(ctx context.Context, asset *sabakan.Asset, remoteAddr string) (bool, error) {
	value, ok := asset.Options[assetOptionAllowedRoles]
	if !ok {
		return true, nil
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false, fmt.Errorf("invalid remote address %q: %w", remoteAddr, err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false, nil
	}

	machines, err := s.Model.Machine.Query(ctx, sabakan.Query{"ipv4": ip.String()})
	if err != nil {
		return false, err
	}
	roles := strings.Split(value, ",")
	for _, m := range machines {
		if slices.Contains(roles, m.Spec.Role) {
			return true, nil
		}
	}
	return false, nil
}
