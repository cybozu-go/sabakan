package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybozu-go/sabakan/models/mock"
)

func TestIgnitions(t *testing.T) {
	// TODO test igitions
}

func testIgnitionTemplatesGet(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/ignitions/cs", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	_, err := m.Ignition.PutTemplate(context.Background(), "cs", "hoge")
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Ignition.PutTemplate(context.Background(), "cs", "fuga")
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/ignitions/cs", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
	var data []string
	err = json.NewDecoder(resp.Body).Decode(&data)
	if len(data) != 2 {
		t.Error("len(data) != 2:", len(data))
	}
}

func testIgnitionTemplatesPut(t *testing.T) {
	t.Parallel()

	ign := `{ "ignition": { "version": "2.2.0" } }`

	m := mock.NewModel()
	handler := Server{Model: m}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/ignitions/cs", bytes.NewBufferString(ign))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Error("resp.StatusCode != http.StatusCreated:", resp.StatusCode)
	}

	tmpl, err := m.Ignition.GetTemplate(context.Background(), "cs", "0")
	if err != nil {
		t.Fatal(err)
	}
	if tmpl != ign {
		t.Error("unexpected template:", tmpl)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/", bytes.NewBufferString(ign))
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	// TODO check if the ignition is valid
}

func testIgnitionTemplatesDelete(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	_, err := m.Ignition.PutTemplate(context.Background(), "cs", "hello")
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/v1/ignitions/cs/0", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/ignitions/cs/99", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/ignitions//", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}
}

func TestIgnitionTemplatess(t *testing.T) {
	t.Run("TemplateGet", testIgnitionTemplatesGet)
	t.Run("TemplatePut", testIgnitionTemplatesPut)
	t.Run("TemplateDelete", testIgnitionTemplatesDelete)
}
