package sabakan

import (
	"encoding/json"
	"net/http"

	"github.com/cybozu-go/log"
)

func respWriter(w http.ResponseWriter, data interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}
	return nil
}

func respError(w http.ResponseWriter, resperr error, status int) {
	out := map[string]interface{}{
		"error": resperr.Error(),
	}
	log.Error(resperr.Error(), nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		log.Error(err.Error(), nil)
	}
}
