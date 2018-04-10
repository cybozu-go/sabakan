package sabakan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func Test_respError(t *testing.T) {
	w := httptest.NewRecorder()
	resperr := fmt.Errorf("test")
	respError(w, resperr, http.StatusBadRequest)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("expected 400. actual: ", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatal("expected application/json")
	}
	expected := "{\"error\":\"test\"}"
	if string(body) != expected {
		t.Fatal("invalid response body, ", string(expected))
	}
}
