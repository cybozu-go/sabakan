package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybozu-go/sabakan/v3"
	"github.com/cybozu-go/sabakan/v3/models/mock"
)

func TestHandleVersion(t *testing.T) {
	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/version", nil)
	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("failed to read response body", err)
	}

	body := make(map[string]string)
	err = json.Unmarshal(b, &body)
	if err != nil {
		t.Fatal("failed to unmarshal response body", err)
	}

	if body["version"] != sabakan.Version {
		t.Fatal("body[\"version\"] != sabakan.Version:", body["version"])
	}
}
