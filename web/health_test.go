package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybozu-go/sabakan/v3/models/mock"
)

func TestHandleHealth(t *testing.T) {
	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatuCode != http.StatusOK:", resp.StatusCode)
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

	if body["health"] != "healthy" {
		t.Fatal("body[\"health\"] != \"healthy\":", body["health"])
	}
}
