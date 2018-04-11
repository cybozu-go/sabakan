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
	crypt := sabakanCrypt{Path: "path", Key: "aaa"}
	respWriter(w, crypt, http.StatusCreated)
	resp := w.Result()
	var sut sabakanCrypt
	json.NewDecoder(resp.Body).Decode(&sut)

	if resp.StatusCode != http.StatusCreated {
		t.Fatal("expected 201. actual: ", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/json" {
		t.Fatal("expected application/json")
	}
	if sut != crypt {
		t.Fatal("invalid response body")
	}
}

func TestRespError(t *testing.T) {
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
	expected := "{\"error\":\"test\"}\n"
	if string(body) != expected {
		t.Fatal("actual:", string(body), ", expected:", expected)
	}
}
