package client

import (
	"encoding/json"
	"os"
	"testing"

	ign23 "github.com/coreos/ignition/config/v2_3/types"
	"github.com/google/go-cmp/cmp"
	"github.com/vincent-petithory/dataurl"
)

func TestBuildIgnitionTemplate(t *testing.T) {
	t.Run("2.3", testBuildIgnitionTemplate2_3)
}

func testBuildIgnitionTemplate2_3(t *testing.T) {
	t.Parallel()

	meta := map[string]interface{}{
		"foo": []int{1, 2, 3},
		"bar": "zot",
	}
	tmpl, err := BuildIgnitionTemplate("../testdata/test/test.yml", meta)
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.Version != Ignition2_3 {
		t.Error(`tmpl.Version != Ignition2_3:`, tmpl.Version)
	}
	if !cmp.Equal(meta, tmpl.Metadata) {
		t.Error("wrong meta data:", cmp.Diff(meta, tmpl.Metadata))
	}

	var cfg ign23.Config
	err = json.Unmarshal(tmpl.Template, &cfg)
	if err != nil {
		t.Fatal(err)
	}

	boolPtr := func(b bool) *bool { return &b }
	intPtr := func(i int) *int { return &i }
	strPtr := func(s string) *string { return &s }
	expected := ign23.Config{}
	expected.Passwd.Groups = []ign23.PasswdGroup{
		{
			Name: "cybozu",
			Gid:  intPtr(10000),
		},
	}
	expected.Passwd.Users = []ign23.PasswdUser{
		{
			Name:              "core",
			PasswordHash:      strPtr("$6$43y3tkl..."),
			SSHAuthorizedKeys: []ign23.SSHAuthorizedKey{"key1"},
		},
	}
	fi0, err := os.Stat("../testdata/base/files/etc/rack")
	if err != nil {
		t.Fatal(err)
	}
	fi3, err := os.Stat("../testdata/test/files/etc/hostname")
	if err != nil {
		t.Fatal(err)
	}
	expectedFiles := make([]ign23.File, 4)
	expectedFiles[0].Filesystem = "root"
	expectedFiles[0].Path = "/etc/rack"
	expectedFiles[0].Contents.Source = "data:," + dataurl.EscapeString("{{ .Spec.Rack }}\n")
	expectedFiles[0].Mode = intPtr(int(fi0.Mode().Perm()))
	expectedFiles[1].Filesystem = "root"
	expectedFiles[1].Path = "/tmp/foo.img"
	expectedFiles[1].Contents.Source = "{{ MyURL }}/api/v1/assets/foo.img"
	expectedFiles[2].Filesystem = "root"
	expectedFiles[2].Path = "/opt/sbin/bar"
	expectedFiles[2].Contents.Source = "{{ MyURL }}/api/v1/assets/bar"
	expectedFiles[2].Mode = intPtr(0755)
	expectedFiles[3].Filesystem = "root"
	expectedFiles[3].Path = "/etc/hostname"
	expectedFiles[3].Contents.Source = "data:," + dataurl.EscapeString("{{ .Spec.Serial }}\n")
	expectedFiles[3].Mode = intPtr(int(fi3.Mode().Perm()))
	expected.Storage.Files = expectedFiles
	expected.Networkd.Units = []ign23.Networkdunit{
		{
			Name: "10-node0.netdev",
			Contents: `[NetDev]
Name=node0
Kind=dummy
Address={{ index .Spec.IPv4 0 }}/32
`,
		},
	}
	expected.Systemd.Units = []ign23.Unit{
		{
			Name:    "chronyd.service",
			Enabled: boolPtr(true),
			Contents: `[Unit]
Description=Chrony

[Service]
ExecStart=/usr/bin/chronyd

[Install]
WantedBy=multi-user.target
`,
		},
		{
			Name:     "bird.service",
			Contents: "[Unit]\nDescription=bird\n",
		},
		{
			Name: "update-engine.service",
			Mask: true,
		},
	}
	if !cmp.Equal(expected, cfg) {
		t.Error("unexpected build result:", cmp.Diff(expected, cfg))
	}
}
