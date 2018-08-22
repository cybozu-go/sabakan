package web

import "net/http"

func (s Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		hs, err := s.Model.Health.GetHealth(r.Context())
		if err != nil {
			renderJSON(w, hs, http.StatusInternalServerError)
			return
		}

		renderJSON(w, hs, http.StatusOK)
	}
}
