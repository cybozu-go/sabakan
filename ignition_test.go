package sabakan

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

func TestValidateIgnitionTemplate(t *testing.T) {
	testIPAMConfig := &IPAMConfig{
		MaxNodesInRack:  28,
		NodeIPv4Pool:    "10.69.0.0/20",
		NodeIPv4Offset:  "",
		NodeRangeSize:   6,
		NodeRangeMask:   26,
		NodeIndexOffset: 3,
		NodeIPPerNode:   3,
		BMCIPv4Pool:     "10.72.16.0/20",
		BMCIPv4Offset:   "0.0.1.0",
		BMCRangeSize:    5,
		BMCRangeMask:    20,
	}

	tmpls := []string{`{ "ignition": { "version": "2.3.0" } }`,
		`{
		  "ignition": { "version": "2.3.0" },
		  "storage": {
		    "files": [{
		      "filesystem": "root",
		      "path": "/etc/hostname",
		      "mode": 420,
		      "contents": {
		        "source": "{{.Serial}}"
		      }
		    }]
		  }
		}`,
	}
	for _, tmpl := range tmpls {
		err := ValidateIgnitionTemplate(tmpl, nil, testIPAMConfig)
		if err != nil {
			t.Errorf("err != nil: %v, %s", err, tmpl)
		}
	}

	tmpls = []string{
		`{ "ignition" : "" }`,
		`{}`,
		`{
		  "ignition": { "version": "2.3.0',
		  "storage" : {
		    "files": [{
		      "filesystem": "root",
		      "path": "/etc/hostname"
		      "mode": 420
		      "contents": {
		        "source": "{{.User}}"
		      }
		    }]
		  }
		}`,
	}

	for _, tmpl := range tmpls {
		err := ValidateIgnitionTemplate(tmpl, nil, testIPAMConfig)
		if err == nil {
			t.Error("err == nil:", err)
		}
	}
}

func TestRenderIgnition(t *testing.T) {
	cases := []struct {
		tmpl string
		mc   *Machine
		ign  string
	}{
		{
			` { "ignition": { "version": "2.3.0" } }`,
			&Machine{},
			`{"ignition":{"version":"2.3.0"}}`,
		}, {
			`{
			  "ignition": { "version": "2.3.0" },
			  "storage": {
			    "files": [
			      {
			        "path": "/opt/file1",
			        "filesystem": "root",
			        "contents": { "source": "{{.Serial}}" },
			        "mode": 420,
			        "user": { "id": 500 },
			        "group": { "id": 501 }
			      }, {
			        "path": "/etc/ignitions/version",
			        "contents": { "source": "{{ Metadata \"version\" }}" }
			      }
			    ]
			  }
			}`,
			NewMachine(MachineSpec{Serial: "abcd, 1234"}),
			`{"ignition":{"version":"2.3.0"},"storage":{"files":[{"filesystem":"root","group":{"id":501},"path":"/opt/file1","user":{"id":500},"contents":{"source":"data:,abcd%2C%201234"},"mode":420},{"contents":{"source":"data:,20181010"},"path":"/etc/ignitions/version"}]}}`},
	}
	u, err := url.Parse("http://localhost:10080")
	if err != nil {
		t.Fatal(err)
	}

	metadata := map[string]string{
		"version": "20181010",
	}
	for _, c := range cases {
		ign, err := RenderIgnition(c.tmpl, &IgnitionParams{Metadata: metadata, MachineSpec: c.mc.Spec, MyURL: u})
		if err != nil {
			t.Fatal(err)
		}
		var expected map[string]interface{}
		var actual map[string]interface{}
		err = json.Unmarshal([]byte(c.ign), &expected)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal([]byte(ign), &actual)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Error("unexpected ignitions:", ign)
		}
	}
}
