package sabakan

import "testing"

func TestValidateIgnitionTemplate(t *testing.T) {
	tmpls := []string{
		`{ "ignition": { "version": "2.2.0" } }`,
		`{ "ignition": { "version": "2.2.0" },
		  "storage": { "files": [{ "filesystem": "root", "path": "/etc/hostname", "mode": 420, "contents": { "source": "data:,{{.Serial}}" } }] } }`,
	}
	for _, tmpl := range tmpls {
		err := ValidateIgnitionTemplate(tmpl)
		if err != nil {
			t.Error("err != nil:", err)
		}
	}

	tmpls = []string{
		`{ "ignition": { "version": "0.1.0" } }`,
		`{ "ignition": { "version": "2.2.0`,
		`{}`,
		`{ "ignition": { "version": "2.2.0" },
		  "storage": { "files": [{ "filesystem": "root", "path": "/etc/hostname", "mode": 420, "contents": { "source": "data:,{{.User}}" } }] } }`,
	}
	for _, tmpl := range tmpls {
		err := ValidateIgnitionTemplate(tmpl)
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

		{`{ "ignition": { "version": "2.2.0" } }`, &Machine{}, `{ "ignition": { "version": "2.2.0" } }`},
		{`{ "ignition": { "version": "2.2.0" },
"storage": { "files": [{ "filesystem": "root", "path": "/etc/hostname", "mode": 420, "contents": { "source": "data:,{{.Serial}}" } }] } }`,
			&Machine{Serial: "abcd1234"},
			`{ "ignition": { "version": "2.2.0" },
"storage": { "files": [{ "filesystem": "root", "path": "/etc/hostname", "mode": 420, "contents": { "source": "data:,abcd1234" } }] } }`},
	}
	for _, c := range cases {
		ign, err := RenderIgnition(c.tmpl, c.mc)
		if err != nil {
			t.Fatal(err)
		}
		if ign != c.ign {
			t.Error("unexpected ignitions:", ign)
		}

	}
}
