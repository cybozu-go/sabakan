package sabakan

import (
	"encoding/json"
	"net/http"

	"fmt"

	"strings"

	"strconv"

	"github.com/cybozu-go/log"
)

func respWriter(w http.ResponseWriter, data interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	out, err := json.Marshal(data)
	outstr := string(out)
	log.Info(strings.Join([]string{"status:", strconv.Itoa(status), ", response_body:", outstr}, " "), nil)
	fmt.Fprintf(w, outstr)
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
