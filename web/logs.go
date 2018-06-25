package web

import (
	"net/http"
	"time"
)

func parseDate(v string) (time.Time, error) {
	if len(v) == 0 {
		return time.Time{}, nil
	}

	return time.Parse("20060102", v)
}

func (s Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		renderError(r.Context(), w, APIErrBadMethod)
		return
	}

	since, err := parseDate(r.FormValue("since"))
	if err != nil {
		renderError(r.Context(), w, BadRequest(err.Error()))
		return
	}
	until, err := parseDate(r.FormValue("until"))
	if err != nil {
		renderError(r.Context(), w, BadRequest(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = s.Model.Log.Dump(r.Context(), since, until, w)
	if err != nil {
		renderError(r.Context(), w, InternalServerError(err))
	}
}
