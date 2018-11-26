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
	handler := newTestServer(m)

	cases := []struct {
		machine  string
		expected int
	}{
		{`[{
  "serial": "1234abcd",
  "labels": {
	  "product": "R630",
	  "datacenter": "ty3"
  },
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusCreated},
		{`[{
  "serial": "1234abcd",
  "labels": {
	  "product": "R630",
	  "datacenter": "ty3"
  },
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusConflict},
		{`[{
  "serial": "1234abcd",
  "labels": {
	  "product": "R630"
  },
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusConflict},
		{`[{
  "labels": {
	  "product": "R630",
	  "datacenter": "ty3"
  },
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "5678abcd",
  "labels": {
	  "datacenter": "ty3"
  },
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusCreated},
		{`[{
  "serial": "0000abcd",
  "labels": {
  	  "product": "R630"
  },
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusCreated},
		{`[{
  "serial": "2222abcd",
  "labels": {
	  "product": "R630",
	  "datacenter": "ty3"
  },
  "rack": 1,
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "2222abcd",
  "labels": {
	  "product": "R630",
	  "datacenter": "ty3"
  },
  "rack": 1,
  "role": "boot"
}]`, http.StatusBadRequest},
		{`[{
  "serial": "2222abcd",
  "labels": {
	  "product": "R630",
	  "datacenter": "ty3"
  },
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "&invalid$ bmc#type+"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "2222abcd",
  "labels": {
	  "product": "R630",
	  "datacenter": "ty3"
  },
  "rack": 1,
  "role": "invalid/Role",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "2222abcd",
  "labels": {
	  "invalidKey~+ = ';": "invalidkey"
  },
  "rack": 1,
  "role": "invalid/Role",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "2222abcd",
  "labels": {
	  "invalid_value": "invalid\n\tvalue"
  },
  "rack": 1,
  "role": "invalid/Role",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusBadRequest},
		{`[{
  "serial": "3333abcd",
  "labels": {},
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusCreated},
		{`[{
  "serial": "3333efgh",
  "rack": 1,
  "role": "boot",
  "bmc": {"type": "iDRAC-9"}
}]`, http.StatusCreated},
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

		_, err := m.Machine.Get(context.Background(), "1234abcd")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testMachinesGet(t *testing.T) {
	m := mock.NewModel()
	handler := Server{Model: m}

	m.Machine.Register(context.Background(), []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "1234abcd",
			Labels: map[string]string{
				"product":    "R630",
				"datacenter": "ty3",
			},
			Rack: 1,
			Role: "boot",
			BMC:  sabakan.MachineBMC{Type: "iDRAC-9"},
		}),
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "5678abcd",
			Labels: map[string]string{
				"product":    "R740",
				"datacenter": "ty3",
			},
			Rack: 1,
			Role: "worker",
			BMC:  sabakan.MachineBMC{Type: "iDRAC-9"},
		}),
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "1234efgh",
			Labels: map[string]string{
				"product":    "R630",
				"datacenter": "ty3",
			},
			Rack: 2,
			Role: "boot",
			BMC:  sabakan.MachineBMC{Type: "IPMI-2.0"},
		}),
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
			query:    map[string][]string{"labels": {"product=R630"}},
			status:   http.StatusOK,
			expected: map[string]bool{"1234abcd": true, "1234efgh": true},
		},
		{
			query:    map[string][]string{"labels": {"datacenter=ty3"}},
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
			query:    map[string][]string{"bmc-type": {"iDRAC-9"}},
			status:   http.StatusOK,
			expected: map[string]bool{"1234abcd": true, "5678abcd": true},
		},
		{
			query:    map[string][]string{"state": {"uninitialized"}},
			status:   http.StatusOK,
			expected: map[string]bool{"1234abcd": true, "5678abcd": true, "1234efgh": true},
		},

		{
			query:    map[string][]string{"serial": {"5689abcd"}},
			status:   http.StatusNotFound,
			expected: nil,
		},
		{
			query:    map[string][]string{"state": {"unreachable"}},
			status:   http.StatusNotFound,
			expected: nil,
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
			if m.Status.Duration == 0 {
				t.Error("duration should not be zero")
			}
			serials[m.Spec.Serial] = true
		}
		if !reflect.DeepEqual(serials, c.expected) {
			t.Errorf("wrong query result: %#v", serials)
		}
	}
}

func testMachinesDelete(t *testing.T) {
	m := mock.NewModel()
	handler := newTestServer(m)

	m1 := sabakan.NewMachine(sabakan.MachineSpec{
		Serial: "1234abcd",
		Labels: map[string]string{
			"product":    "R630",
			"datacenter": "ty3",
		},
		Rack: 1,
		Role: "boot",
	})
	m2 := sabakan.NewMachine(sabakan.MachineSpec{
		Serial: "qqq",
		Labels: map[string]string{
			"product":    "R630",
			"datacenter": "ty3",
		},
		Rack: 2,
		Role: "cs",
	})
	m2.Status.State = sabakan.StateRetired

	err := m.Machine.Register(context.Background(), []*sabakan.Machine{m1, m2})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		serial string
		status int
	}{
		{
			serial: "1234abcd",
			status: http.StatusInternalServerError,
		},
		{
			serial: "qqq",
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

func testMachinesGraphQL(t *testing.T) {
	m := mock.NewModel()
	handler := NewServer(m, "", nil, nil, false)

	m.Machine.Register(context.Background(), []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "1234abcd",
			Labels: map[string]string{
				"product":    "R630",
				"datacenter": "ty3",
			},
			Rack: 1,
			Role: "boot",
			BMC:  sabakan.MachineBMC{Type: "iDRAC-9"},
		}),
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "5678abcd",
			Labels: map[string]string{
				"product":    "R740",
				"datacenter": "ty3",
			},
			Rack: 1,
			Role: "worker",
			BMC:  sabakan.MachineBMC{Type: "iDRAC-9"},
		}),
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "1234efgh",
			Labels: map[string]string{
				"product":    "R630",
				"datacenter": "ty3",
			},
			Rack: 2,
			Role: "boot",
			BMC:  sabakan.MachineBMC{Type: "IPMI-2.0"},
		}),
	})

	v := url.Values{}
	v.Set("query", `{searchMachines { spec { serial } } }`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/graphql?"+v.Encode(), nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("wrong status code:", resp.StatusCode)
	}

	var gqlResponse struct {
		Errors []interface{} `json:"errors"`
		Data   struct {
			SearchMachines []interface{} `json:"searchMachines"`
		} `json:"data"`
	}
	err := json.NewDecoder(resp.Body).Decode(&gqlResponse)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if len(gqlResponse.Data.SearchMachines) != 3 {
		t.Error(`len(gqlResponse.Data.SearchMachines) != 3`, gqlResponse)
	}
}

func TestMachines(t *testing.T) {
	t.Run("Get", testMachinesGet)
	t.Run("Post", testMachinesPost)
	t.Run("Delete", testMachinesDelete)
	t.Run("GraphQL", testMachinesGraphQL)
}
