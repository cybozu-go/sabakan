package sabakan

import (
	"reflect"
	"testing"
)

func TestValidateIgnitionTemplate(t *testing.T) {
	testIPAMConfig = &IPAMConfig{
		MaxNodesInRack:  28,
		NodeIPv4Pool:    "10.69.0.0/20",
		NodeRangeSize:   6,
		NodeRangeMask:   26,
		NodeIndexOffset: 3,
		NodeIPPerNode:   3,
		BMCIPv4Pool:     "10.72.17.0/20",
		BMCRangeSize:    5,
		BMCRangeMask:    20,
	}

	tmpls := []string{
		`ignition:`,
		`storage:
  files:
  - filesystem: root
    path: "/etc/hostname"
    mode: 420
    contents:
      inline: "{{.Serial}}"`,
	}
	for _, tmpl := range tmpls {
		err := ValidateIgnitionTemplate(tmpl, testIPAMConfig)
		if err != nil {
			t.Error("err != nil:", err)
		}
	}

	tmpls = []string{
		`ignition`,
		``,
		`storage:
  files:
  - filesystem: root
    path: "/etc/hostname"
    mode: 420
    contents:
      inline: "{{.User}}"`,
	}

	for _, tmpl := range tmpls {
		err := ValidateIgnitionTemplate(tmpl, testIPAMConfig)
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
		{`ignition:`, &Machine{}, `{"ignition":{"config":{},"security":{"tls":{}},"timeouts":{},"version":"2.2.0"},"networkd":{},"passwd":{},"storage":{},"systemd":{}}`},
		{`storage:
  files:
    - path: /opt/file1
      filesystem: root
      contents:
        inline: {{.Serial}}
      mode: 0644
      user:
        id: 500
      group:
        id: 501`,
			&Machine{Serial: "abcd1234"},
			`{"ignition":{"config":{},"security":{"tls":{}},"timeouts":{},"version":"2.2.0"},"networkd":{},"passwd":{},"storage":{"files":[{"filesystem":"root","group":{"id":501},"path":"/opt/file1","user":{"id":500},"contents":{"source":"data:,abcd1234","verification":{}},"mode":420}]},"systemd":{}}`},
	}
	for _, c := range cases {
		ign, err := RenderIgnition(c.tmpl, c.mc)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(ign, c.ign) {
			t.Error("unexpected ignitions:", ign)
		}

	}
}
