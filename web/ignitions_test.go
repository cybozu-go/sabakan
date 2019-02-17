package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
	"github.com/google/go-cmp/cmp"
)

func TestIgnitionTemplates(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	testWithIPAM(t, m)
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/ignitions/cs/1.0.0", nil)
	handler.ServeHTTP(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	testIgn := `{"ignition":{"version":"2.3.0"}}`
	tmpl := &sabakan.IgnitionTemplate{
		Version:  sabakan.Ignition2_3,
		Template: json.RawMessage(testIgn),
		Metadata: map[string]interface{}{"foo": "bar"},
	}
	data, err := json.Marshal(tmpl)
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/cs/1.0.0", bytes.NewReader(data))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Error("resp.StatusCode != http.StatusCreated:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/cs/1.0.0", bytes.NewReader(data))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusConflict {
		t.Error("resp.StatusCode != http.StatusConflict:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/bad%20role/1.0.0", bytes.NewReader(data))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/cs/bad.version", bytes.NewReader(data))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/cs/1.0.0-rc1", bytes.NewReader(data))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Error("resp.StatusCode != http.StatusCreated:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/cs/1.1.0", bytes.NewReader(data))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Error("resp.StatusCode != http.StatusCreated:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/ignitions/cs", nil)
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	} else {
		var ids []string
		if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
			t.Fatal(err)
		}
		expected := []string{"1.0.0-rc1", "1.0.0", "1.1.0"}
		if !cmp.Equal(expected, ids) {
			t.Error("wrong ids:", cmp.Diff(expected, ids))
		}
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/ignitions/cs/1.0.0", nil)
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
	tmpl2 := new(sabakan.IgnitionTemplate)
	if err := json.NewDecoder(resp.Body).Decode(tmpl2); err != nil {
		t.Fatal(err)
	}
	if tmpl.Version != tmpl2.Version {
		t.Error("wrong template version:", tmpl2.Version)
	}
	if testIgn != string(tmpl2.Template) {
		t.Error("wrong template ignition:", string(tmpl2.Template))
	}
	if !cmp.Equal(tmpl.Metadata, tmpl2.Metadata) {
		t.Error("wrong template meta data:", cmp.Diff(tmpl.Metadata, tmpl2.Metadata))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/ignitions/cs/1.0.0", nil)
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/ignitions/cs/1.0.0", nil)
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/ignitions/cs/1.0.0", nil)
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}
}
