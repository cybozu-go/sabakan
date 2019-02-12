package web

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"reflect"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
)

func TestIgnitions(t *testing.T) {
	t.Parallel()

	ign := `{
"ignition": { "version": "2.2.0" },
  "storage": {
    "files": [
      {
        "filesystem": "root",
        "path": "/etc/hostname",
        "mode": 420,
        "contents": { "source": "{{.Spec.Serial}}" }
      }, {
        "filesystem": "root",
        "mode": 420,
        "path": "/etc/neco/rack",
	"contents": { "source": "{{ .Spec.Rack }}" }
      }
    ]
  },
  "networkd": {
    "units": [
      {
        "name": "10-node0.network",
        "contents": "[Match]\nName=node0\n\n[Network]\nAddress={{ index .Spec.IPv4 0 }}/32\n"
      }
    ]
  }
}
`

	expected := `{
  "ignition": {
    "version": "2.2.0"
  },
  "networkd": {
    "units": [
      {
        "contents": "[Match]\nName=node0\n\n[Network]\nAddress=10.69.0.4/32\n",
        "name": "10-node0.network"
      }
    ]
  },
  "storage": {
    "files": [
      {
        "filesystem": "root",
        "path": "/etc/hostname",
        "contents": {
          "source": "data:,2222abcd"
        },
        "mode": 420
      },
      {
        "filesystem": "root",
        "path": "/etc/neco/rack",
        "contents": {
          "source": "data:,1"
        },
        "mode": 420
      }
    ]
  }
}
`

	m := mock.NewModel()
	handler := Server{Model: m}

	err := m.Ignition.PutTemplate(context.Background(), "cs", "1.0.0", ign, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	machines := []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "2222abcd",
			Labels: map[string]string{
				"product":    "R630",
				"datacenter": "ty3",
			},
			Rack: 1,
			Role: "cs",
			IPv4: []string{"10.69.0.4"},
			BMC:  sabakan.MachineBMC{Type: "iDRAC-9"},
		}),
	}

	err = m.Machine.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/boot/ignitions/2222abcd/1.0.0", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var d1, d2 map[string]interface{}
	err = json.Unmarshal([]byte(body), &d1)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal([]byte(expected), &d2)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(d1, d2) {
		t.Errorf("unexpected ignition actual:%v, expected:%v", d1, d2)
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

func testIgnitionTemplateMetadataGet(t *testing.T) {
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

	err := m.Ignition.PutTemplate(context.Background(), "cs", "1.12.34", "hoge", map[string]string{"version": "1.12.34"})
	if err != nil {
		t.Fatal(err)
	}
	err = m.Ignition.PutTemplate(context.Background(), "cs", "1.2.3", "fuga", map[string]string{"version": "1.2.3"})
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
	var data []sabakan.IgnitionInfo
	err = json.NewDecoder(resp.Body).Decode(&data)
	if len(data) != 2 {
		t.Error("len(data) != 2:", len(data))
	}
	if data[0].ID != "1.2.3" {
		t.Error("data[0].ID != 1.2.3")
	}
	if data[1].ID != "1.12.34" {
		t.Error("data[1].ID != 1.12.34")
	}
	if data[0].Metadata["version"] != "1.2.3" {
		t.Error("wrong version: ", data[0])
	}
	if data[1].Metadata["version"] != "1.12.34" {
		t.Error("wrong version: ", data[1])
	}
}

func testIgnitionTemplatesGet(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/ignitions/cs/1.0.0", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	err := m.Ignition.PutTemplate(context.Background(), "cs", "1.0.0", "hoge", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/ignitions/cs/1.0.0", nil)
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

func testIgnitionTemplatesPost(t *testing.T) {
	t.Parallel()

	ign := `{ "ignition" : { "version": "2.2.0" } }`
	invalid := `{ "ignition" : {} }`

	m := mock.NewModel()
	handler := newTestServer(m)

	config := &sabakan.IPAMConfig{
		MaxNodesInRack:    28,
		NodeIPv4Pool:      "10.69.0.0/20",
		NodeIPv4Offset:    "",
		NodeRangeSize:     6,
		NodeRangeMask:     26,
		NodeIPPerNode:     3,
		NodeIndexOffset:   3,
		NodeGatewayOffset: 1,
		BMCIPv4Pool:       "10.72.16.0/20",
		BMCIPv4Offset:     "0.0.1.0",
		BMCRangeSize:      5,
		BMCRangeMask:      20,
		BMCGatewayOffset:  1,
	}

	err := m.IPAM.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/ignitions/cs/1.0.0", bytes.NewBufferString(ign))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Error("resp.StatusCode != http.StatusCreated:", resp.StatusCode)
	}

	tmpl, err := m.Ignition.GetTemplate(context.Background(), "cs", "1.0.0")
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
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/cs", bytes.NewBufferString(ign))
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/cs/1.0.0", bytes.NewBufferString(invalid))
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/@invalidRole/1.0.0", bytes.NewBufferString(ign))
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/ignitions/cs/a-b.c-d", bytes.NewBufferString(ign))
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}
}

func testIgnitionTemplatesDelete(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	err := m.Ignition.PutTemplate(context.Background(), "cs", "1.0.0", "hello", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/v1/ignitions/cs/1.0.0", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/ignitions/cs/1.2.0", nil)
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
	t.Run("TemplateMetadataGet", testIgnitionTemplateMetadataGet)
	t.Run("TemplatesGet", testIgnitionTemplatesGet)
	t.Run("TemplatePost", testIgnitionTemplatesPost)
	t.Run("TemplateDelete", testIgnitionTemplatesDelete)
}
