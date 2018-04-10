package sabakan

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_respWriter(t *testing.T) {
	w := httptest.NewRecorder()
	entity := cryptEntity{Path: "path", Key: "aaa"}
	respWriter(w, entity, http.StatusCreated)
	resp := w.Result()
	var sut cryptEntity
	json.NewDecoder(resp.Body).Decode(&sut)

	if resp.StatusCode != http.StatusCreated {
		t.Fatal("expected 201. actual: ", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatal("expected application/json")
	}
	if sut != entity {
		t.Fatal("invalid response body")
	}
}
