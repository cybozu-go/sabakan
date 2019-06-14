package web

import (
	"net/http"
)

func (s Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		err := s.Model.Health.GetHealth(r.Context())

		if err == nil {
			healthStatus := map[string]string{"health": "healthy"}
			renderJSON(w, healthStatus, http.StatusOK)
			return
		}

		healthStatus := map[string]string{"health": "unhealthy"}
		renderJSON(w, healthStatus, http.StatusInternalServerError)
		return
	}

	renderError(r.Context(), w, APIErrBadMethod)
}
