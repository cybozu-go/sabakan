package sabakan

import (
	"encoding/json"
	"net/http"

	"github.com/cybozu-go/log"
)

func renderJSON(w http.ResponseWriter, data interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	return err
}

func renderError(w http.ResponseWriter, resperr error, status int) {
	out := map[string]interface{}{
		"error": resperr.Error(),
	}
	log.Error("an error occurred during request processing", map[string]interface{}{
		"error": resperr.Error(),
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		log.Error("failed to encode error to JSON", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
