package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ign23 "github.com/coreos/ignition/config/v2_3/types"
	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
	"github.com/google/go-cmp/cmp"
)

func TestIgnitions(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	ctx := context.Background()
	testWithIPAM(t, m)
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/boot/ignitions/abc/1.0.0", nil)
	handler.ServeHTTP(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	mc := sabakan.NewMachine(sabakan.MachineSpec{
		Serial: "abc",
		Role:   "cs",
		Rack:   1,
	})
	err := m.Machine.Register(ctx, []*sabakan.Machine{mc})
	if err != nil {
		t.Fatal(err)
	}

	tmpl := &sabakan.IgnitionTemplate{
		Version:  sabakan.Ignition2_3,
		Template: json.RawMessage(`{}`),
	}
	err = m.Ignition.PutTemplate(ctx, "cs", "1.0.0", tmpl)
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/boot/ignitions/abc/1.0.0", nil)
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
	ign := new(ign23.Config)
	if err := json.NewDecoder(resp.Body).Decode(ign); err != nil {
		t.Fatal(err)
	}

	if ign.Ignition.Version != "2.3.0" {
		t.Error(`ign.Ignition.Version != "2.3.0"`)
	}
}
func TestRenderIgnition(t *testing.T) {
	t.Run("2.3", testRenderIgnition2_3)
}

func testRenderIgnition2_3(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	ipam := testWithIPAM(t, m)
	mc := sabakan.NewMachine(sabakan.MachineSpec{
		Serial:      "abc",
		Role:        "cs",
		Rack:        1,
		IndexInRack: 4,
	})
	ipam.GenerateIP(mc)
	strPtr := func(s string) *string { return &s }
	files := make([]ign23.File, 2)
	files[0].Path = "/etc/hostname"
	files[0].Contents.Source = "data:,rack%7B%7B%20.Spec.Rack%20%7D%7D-%7B%7B%20.Spec.Role%20%7D%7D-%7B%7B%20.Spec.IndexInRack%20%7D%7D%0A"
	files[1].Path = "/opt/sbin/sabakan-cryptsetup"
	files[1].Contents.Source = `{{ MyURL }}/api/v1/assets/sabakan-cryptsetup`
	files[1].Contents.Verification.Hash = strPtr(`{{ Metadata "cryptsetuphash" }}`)
	ign := ign23.Config{
		Passwd: ign23.Passwd{
			Groups: []ign23.PasswdGroup{
				{
					Name:         `{{ Metadata "group1" }}`,
					PasswordHash: `{{ Metadata "group1hash" }}`,
				},
			},
			Users: []ign23.PasswdUser{
				{
					Name:         `{{ Metadata "user1" }}`,
					Gecos:        `{{ Metadata "user1gecos" }}`,
					HomeDir:      `{{ Metadata "user1home" }}`,
					Groups:       []ign23.Group{"foo", `{{ "bar" }}`},
					PasswordHash: strPtr(`{{ Metadata "user1hash" }}`),
					PrimaryGroup: `{{ Metadata "user1group" }}`,
					SSHAuthorizedKeys: []ign23.SSHAuthorizedKey{
						`{{ Metadata "user1sshkey" }}`,
					},
					Shell: `{{ Metadata "user1shell" }}`,
				},
			},
		},
		Storage: ign23.Storage{Files: files},
		Networkd: ign23.Networkd{
			Units: []ign23.Networkdunit{
				{
					Name: "10-eno1.network",
					Contents: `[Match]
Name=eno1

[Network]
Address={{ (index .Info.Network.IPv4 0).Address }}/{{ (index .Info.Network.IPv4 0).MaskBits }}
Gateway={{ (index .Info.Network.IPv4 0).Gateway }}
`,
				},
			},
		},
		Systemd: ign23.Systemd{
			Units: []ign23.Unit{
				{
					Name: "foo.service",
					Contents: `[Unit]
Wants=var-lib-foo.mount
After=var-lib-foo.mount

[Service]
ExecStart=/bin/echo {{ add .Spec.Rack 10 }}
`,
				},
			},
		},
	}

	metadata := map[string]interface{}{
		"group1":         "g1",
		"group1hash":     "g1hash",
		"user1":          "u1",
		"user1gecos":     "gegege",
		"user1home":      "/home/u1",
		"user1hash":      "u1hash",
		"user1group":     "u1",
		"user1sshkey":    "u1key",
		"user1shell":     "/bin/bash",
		"cryptsetuphash": "hahaha",
	}

	tmplData, err := json.Marshal(ign)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &sabakan.IgnitionTemplate{
		Version:  sabakan.Ignition2_3,
		Template: json.RawMessage(tmplData),
		Metadata: metadata,
	}

	s := newTestServer(m)
	rendered, err := s.renderIgnition(tmpl, mc)
	if err != nil {
		t.Fatal(err)
	}

	actual, ok := rendered.(*ign23.Config)
	if !ok {
		t.Fatalf("unexpected type: %T", actual)
	}

	// buf := &bytes.Buffer{}
	// enc := json.NewEncoder(buf)
	// enc.SetIndent("", "  ")
	// enc.Encode(rendered)
	// t.Log(buf.String())

	expected := &ign23.Config{}
	expected.Ignition.Version = "2.3.0"
	expected.Passwd.Groups = []ign23.PasswdGroup{
		{
			Name:         "g1",
			PasswordHash: "g1hash",
		},
	}
	expected.Passwd.Users = []ign23.PasswdUser{
		{
			Name:              "u1",
			Gecos:             "gegege",
			HomeDir:           "/home/u1",
			Groups:            []ign23.Group{"foo", "bar"},
			PasswordHash:      strPtr("u1hash"),
			PrimaryGroup:      "u1",
			SSHAuthorizedKeys: []ign23.SSHAuthorizedKey{"u1key"},
			Shell:             "/bin/bash",
		},
	}
	expectedFiles := make([]ign23.File, 2)
	expectedFiles[0].Path = "/etc/hostname"
	expectedFiles[0].Contents.Source = "data:,rack1-cs-4%0A"
	expectedFiles[1].Path = "/opt/sbin/sabakan-cryptsetup"
	expectedFiles[1].Contents.Source = testMyURL + "/api/v1/assets/sabakan-cryptsetup"
	expectedFiles[1].Contents.Verification.Hash = strPtr("hahaha")
	expected.Storage.Files = expectedFiles
	expected.Networkd.Units = []ign23.Networkdunit{
		{
			Name: "10-eno1.network",
			Contents: `[Match]
Name=eno1

[Network]
Address=10.69.0.196/26
Gateway=10.69.0.193
`,
		},
	}
	expected.Systemd.Units = []ign23.Unit{
		{
			Name: "foo.service",
			Contents: `[Unit]
Wants=var-lib-foo.mount
After=var-lib-foo.mount

[Service]
ExecStart=/bin/echo 11
`,
		},
	}
	if !cmp.Equal(expected, actual) {
		t.Error("unexpected ignition:", cmp.Diff(expected, actual))
	}
}
