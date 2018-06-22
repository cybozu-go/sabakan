package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
)

func TestLogs(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	good := `
{
   "max-nodes-in-rack": 28,
   "node-ipv4-pool": "10.69.0.0/20",
   "node-ipv4-range-size": 6,
   "node-ipv4-range-mask": 26,
   "node-index-offset": 3,
   "node-ip-per-node": 3,
   "bmc-ipv4-pool": "10.72.16.0/20",
   "bmc-ipv4-range-size": 5,
   "bmc-ipv4-range-mask": 20
}
`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/config/ipam", strings.NewReader(good))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK")
	}

	r = httptest.NewRequest("GET", "/api/v1/logs", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("request failed with " + http.StatusText(resp.StatusCode))
	}

	dec := json.NewDecoder(resp.Body)
	a := new(sabakan.AuditLog)
	err := dec.Decode(a)
	if err != nil {
		t.Fatal(err)
	}
	if a.Category != sabakan.AuditIPAM {
		t.Error(`a.Category != sabakan.AuditIPAM`, a.Category)
	}

	if dec.More() {
		t.Error(`should have no more JSON`)
	}

	r = httptest.NewRequest("GET", "/api/v1/logs?since=2018-04-04", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error(`resp.StatusCode != http.StatusBadRequest`, resp.StatusCode)
	}

	r = httptest.NewRequest("GET", "/api/v1/logs?until=20190404", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error(`resp.StatusCode != http.StatusOK`, resp.StatusCode)
	}
}
