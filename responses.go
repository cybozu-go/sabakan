package sabakan

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cybozu-go/log"
)

func respError(w http.ResponseWriter, resperr error, status int) {
	out, err := json.Marshal(map[string]interface{}{
		"error": resperr.Error(),
	})
	if err != nil {
		log.Error(err.Error(), nil)
		return
	}

	log.Error(resperr.Error(), nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, string(out))
}
