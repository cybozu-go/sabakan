package sabakan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespWriter(t *testing.T) {
	w := httptest.NewRecorder()
	inputCrypt := sabakanCrypt{Path: "path", Key: "aaa"}
	renderJSON(w, inputCrypt, http.StatusCreated)
	resp := w.Result()
	var outputCrypt sabakanCrypt
	err := json.NewDecoder(resp.Body).Decode(&outputCrypt)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatal("expected 201. actual: ", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatal("expected application/json")
	}
	if outputCrypt != inputCrypt {
		t.Fatal("invalid response body")
	}
}

func TestRespError(t *testing.T) {
	w := httptest.NewRecorder()
	resperr := fmt.Errorf("test")
	renderError(w, resperr, http.StatusBadRequest)
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("expected 400. actual: ", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatal("expected application/json")
	}
	expected := "{\"error\":\"test\"}\n"
	if string(body) != expected {
		t.Fatal("actual:", string(body), ", expected:", expected)
	}
}
