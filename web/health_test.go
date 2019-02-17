package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybozu-go/sabakan/v2/models/mock"
)

func testHandleHealth(t *testing.T) {
	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatuCode != http.StatusOK:", resp.StatusCode)
	}

}
