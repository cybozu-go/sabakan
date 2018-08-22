package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybozu-go/sabakan/models/mock"
)

func testHandleVersion(t *testing.T) {
	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/version", nil)
	handler.ServeHTTP(w, r)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

}
