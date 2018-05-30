package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
)

func testMachinesPost(t *testing.T) {

	m := mock.NewModel()
	handler := Server{Model: m}

	cases := []struct {
		machine  string
		expected int
	}{
		{`[{
  "serial": "1234abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusCreated},
		{`[{
  "serial": "1234abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusConflict},
		{`[{
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "5678abcd",
  "datacenter": "ty3",
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "0000abcd",
  "product": "R630",
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "2222abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "2222abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "role": "boot"
}]`, http.StatusBadRequest},
		{`[{
  "serial": "2222abcd",
  "product": "R630",
  "datacenter": "ty3",
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "unknown-BMC"}
}]`, http.StatusBadRequest},
	}

	for _, c := range cases {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api/v1/machines", strings.NewReader(c.machine))
		handler.ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != c.expected {
			t.Error("wrong status code:", resp.StatusCode)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			continue
		}

		machines, err := m.Machine.Query(context.Background(), sabakan.QueryBySerial("1234abcd"))
		if err != nil {
			t.Fatal(err)
		}
		if len(machines) != 1 {
			t.Error("machine not register")
		}
	}

}

func testMachinesGet(t *testing.T) {
	m := mock.NewModel()
	handler := Server{Model: m}

	m.Machine.Register(context.Background(), []*sabakan.Machine{
		{
			Serial:     "1234abcd",
			Product:    "R630",
			Datacenter: "ty3",
			Rack:       1,
			Role:       "boot",
			BMC:        sabakan.MachineBMC{Type: sabakan.BmcIdrac9},
		},
		{
			Serial:     "5678abcd",
			Product:    "R740",
			Datacenter: "ty3",
			Rack:       1,
			Role:       "worker",
			BMC:        sabakan.MachineBMC{Type: sabakan.BmcIdrac9},
		},
		{
			Serial:     "1234efgh",
			Product:    "R630",
			Datacenter: "ty3",
			Rack:       2,
			Role:       "boot",
			BMC:        sabakan.MachineBMC{Type: sabakan.BmcIpmi2},
		},
	})

	cases := []struct {
		query    url.Values
		status   int
		expected map[string]bool
	}{
		{
			query:    map[string][]string{"serial": {"1234abcd"}},
			status:   http.StatusOK,
			expected: map[string]bool{"1234abcd": true},
		},
		{
			query:    map[string][]string{"product": {"R630"}},
			status:   http.StatusOK,
			expected: map[string]bool{"1234abcd": true, "1234efgh": true},
		},
		{
			query:    map[string][]string{"datacenter": {"ty3"}},
			status:   http.StatusOK,
			expected: map[string]bool{"1234abcd": true, "5678abcd": true, "1234efgh": true},
		},
		{
			query:    map[string][]string{"rack": {"1"}},
			status:   http.StatusOK,
			expected: map[string]bool{"1234abcd": true, "5678abcd": true},
		},
		{
			query:    map[string][]string{"role": {"boot"}},
			status:   http.StatusOK,
			expected: map[string]bool{"1234abcd": true, "1234efgh": true},
		},
		{
			query:    map[string][]string{"serial": {"5689abcd"}},
			status:   http.StatusNotFound,
			expected: nil,
		},
		{
			query:    map[string][]string{"bmc-type": {"iDRAC-9"}},
			status:   http.StatusOK,
			expected: map[string]bool{"1234abcd": true, "5678abcd": true},
		},
	}
	for _, c := range cases {
		w := httptest.NewRecorder()
		u := url.URL{Path: "/api/v1/machines", RawQuery: c.query.Encode()}
		r := httptest.NewRequest("GET", u.String(), nil)

		handler.ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != c.status {
			t.Error("wrong status code:", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusOK {
			continue
		}
		var machines []*sabakan.Machine
		err := json.NewDecoder(resp.Body).Decode(&machines)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		serials := make(map[string]bool)
		for _, m := range machines {
			serials[m.Serial] = true
		}
		if !reflect.DeepEqual(serials, c.expected) {
			t.Errorf("wrong query result: %#v", serials)
		}
	}
}

func testMachinesDelete(t *testing.T) {
	m := mock.NewModel()
	handler := Server{Model: m}

	m.Machine.Register(context.Background(), []*sabakan.Machine{
		{
			Serial:     "1234abcd",
			Product:    "R630",
			Datacenter: "ty3",
			Rack:       1,
			Role:       "boot",
		},
	})

	cases := []struct {
		serial string
		status int
	}{
		{
			serial: "1234abcd",
			status: http.StatusOK,
		},
		{
			serial: "5678efgh",
			status: http.StatusNotFound,
		},
	}
	for _, c := range cases {
		w := httptest.NewRecorder()
		u := path.Join("/api/v1/machines", c.serial)
		r := httptest.NewRequest("DELETE", u, nil)

		handler.ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != c.status {
			t.Error("wrong status code:", resp.StatusCode, c.serial)
		}
	}
}

func TestMachines(t *testing.T) {
	t.Run("Get", testMachinesGet)
	t.Run("Post", testMachinesPost)
	t.Run("Delete", testMachinesDelete)
}
