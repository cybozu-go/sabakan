package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan/models/mock"
)

func TestIgnitions(t *testing.T) {
	t.Parallel()

	ign := `{
    "ignition": { "version": "2.2.0" },
    "storage": {
        "files": [{
            "filesystem": "root",
            "path": "/etc/hostname",
            "mode": 420,
            "contents": { "source": "{{.Serial}}" }
        }]
    }
}`

	expected := `{
    "ignition": { "version": "2.2.0" },
    "storage": {
        "files": [{
            "filesystem": "root",
            "path": "/etc/hostname",
            "mode": 420,
            "contents": { "source": "2222abcd" }
        }]
    }
}`

	m := mock.NewModel()
	handler := Server{Model: m}

	_, err := m.Ignition.PutTemplate(context.Background(), "cs", ign)
	if err != nil {
		t.Fatal(err)
	}

	machine := `[{
  "serial": "2222abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "role": "cs",
  "bmc": {"type": "iDRAC-9"}
}]`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/machines", strings.NewReader(machine))
	handler.ServeHTTP(w, r)

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/boot/ignitions/2222abcd/0", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != expected {
		t.Error("unexpected ignition:", string(body))
	}

	// serial is not found
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/boot/ignitions/1234abcd/0", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	// id is not found
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/boot/ignitions/2222abcd/1", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}
}

func testIgnitionTemplateIDsGet(t *testing.T) {
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

func testIgnitionTemplatesGet(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/ignitions/cs/0", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	_, err := m.Ignition.PutTemplate(context.Background(), "cs", "hoge")
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/ignitions/cs/0", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
	ign, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(ign) != "hoge" {
		t.Error("ign != hoge:", ign)
	}
}

func testIgnitionTemplatesPut(t *testing.T) {
	t.Parallel()

	ign := `{ "ignition": { "version": "2.2.0" } }`
	invalid := `{ "ignition": { "version": "0.2.0" } }`

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

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/cs", bytes.NewBufferString(invalid))
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/@invalidRole", bytes.NewBufferString(ign))
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}
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

func TestIgnitionTemplates(t *testing.T) {
	t.Run("TemplateIDsGet", testIgnitionTemplateIDsGet)
	t.Run("TemplatesGet", testIgnitionTemplatesGet)
	t.Run("TemplatePut", testIgnitionTemplatesPut)
	t.Run("TemplateDelete", testIgnitionTemplatesDelete)
}
