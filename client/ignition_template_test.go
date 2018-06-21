package client

import (
	"reflect"
	"testing"
)

func testIgnitionBuilderConstructIgnitionYAML(t *testing.T) {
	b := newIgnitionBuilder("../testdata/test")

	tests := []struct {
		name    string
		source  *ignitionSource
		wantErr bool
	}{
		{name: "passwd", source: &ignitionSource{Passwd: "../base/passwd.yml"}, wantErr: false},
		{name: "files", source: &ignitionSource{Files: []string{"/etc/hostname"}}, wantErr: false},
		{name: "systemd", source: &ignitionSource{Systemd: []systemd{{Name: "bird.service"}}}, wantErr: false},
		{name: "networkd", source: &ignitionSource{Networkd: []string{"10-node0.netdev"}}, wantErr: false},
		{name: "include yml from different working directory", source: &ignitionSource{Include: "../base/base.yml"}, wantErr: false},

		{name: "passwd not found", source: &ignitionSource{Passwd: "nonexists_file.yml"}, wantErr: true},
		{name: "files not found", source: &ignitionSource{Files: []string{"/etc/not_file"}}, wantErr: true},
		{name: "systemd not found", source: &ignitionSource{Systemd: []systemd{{Name: "hoge.service"}}}, wantErr: true},
		{name: "networkd not found", source: &ignitionSource{Networkd: []string{"nonexists.netdev"}}, wantErr: true},
		{name: "include not found", source: &ignitionSource{Include: "nonexits_base.yml"}, wantErr: true},

		{name: "name not found", source: &ignitionSource{Systemd: []systemd{{Enabled: true}}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := b.constructIgnitionYAML(tt.source); (err != nil) != tt.wantErr {
				t.Errorf("ignitionBuilder.constructIgnitionYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func testIgnitionBuilderConstructFile(t *testing.T) {
	b := newIgnitionBuilder("../testdata/test")
	inputFile := "/etc/hostname"
	if err := b.constructFile(inputFile); err != nil {
		t.Fatal(err)
	}
	if err := b.constructFile(inputFile); err != nil {
		t.Fatal(err)
	}
	storage, ok := b.ignition["storage"].(map[string]interface{})
	if !ok {
		t.Fatal("failed to construct ignition map")
	}

	actual := len(storage["files"].([]interface{}))
	if actual != 2 {
		t.Errorf("Should not overwrite units, so expected length:%d, actual %d", 2, actual)
	}
}

func testIgnitionBuilderConstructSystemd(t *testing.T) {
	b := newIgnitionBuilder("../testdata/test")
	bird := systemd{Name: "bird.service", Enabled: true}
	err := b.constructSystemd(bird)
	if err != nil {
		t.Fatal(err)
	}
	err = b.constructSystemd(bird)
	if err != nil {
		t.Fatal(err)
	}

	updateEngine := systemd{Name: "update-engine.service", Mask: true}
	err = b.constructSystemd(updateEngine)
	if err != nil {
		t.Fatal(err)
	}

	expectedBird := `[Unit]
Description=bird
`
	expected := map[string]interface{}{
		"units": []interface{}{
			map[string]interface{}{
				"name":     "bird.service",
				"enabled":  true,
				"contents": expectedBird,
			},
			map[string]interface{}{
				"name":     "bird.service",
				"enabled":  true,
				"contents": expectedBird,
			},
			map[string]interface{}{
				"name": "update-engine.service",
				"mask": true,
			},
		},
	}

	if !reflect.DeepEqual(b.ignition["systemd"], expected) {
		t.Error("b.ignition['systemd'] is not expected: ", b.ignition["systemd"])
	}
}

func testIgnitionBuilderConstructNetworkd(t *testing.T) {
	b := newIgnitionBuilder("../testdata/test")
	src := "10-node0.netdev"
	err := b.constructNetworkd(src)
	if err != nil {
		t.Fatal(err)
	}
	err = b.constructNetworkd(src)
	if err != nil {
		t.Fatal(err)
	}
	networkd, ok := b.ignition["networkd"].(map[string]interface{})
	if !ok {
		t.Fatal("failed to construct ignition map")
	}

	actual := len(networkd["units"].([]interface{}))
	if actual != 2 {
		t.Errorf("Should not overwrite units, so expected length:%d, actual %d", 2, actual)
	}
}

func testIgnitionBuilderConstructPasswd(t *testing.T) {
	b := newIgnitionBuilder("../testdata/test")
	src := "../base/passwd.yml"
	err := b.constructPasswd(src)
	if err != nil {
		t.Fatal(err)
	}
	err = b.constructPasswd(src)
	if err != nil {
		t.Fatal(err)
	}
	passwd, ok := b.ignition["passwd"].(map[interface{}]interface{})
	if !ok {
		t.Fatal("failed to construct ignition map")
	}

	actual := len(passwd["users"].([]interface{}))
	if actual != 1 {
		t.Errorf("constructPasswd does not append unit, so expected length:%d, actual %d", 1, actual)
	}
}

func testIgnitionBuilderConstructInclude(t *testing.T) {
	b := newIgnitionBuilder("../testdata/test")

	err := b.constructIgnitionYAML(&ignitionSource{Systemd: []systemd{{Name: "bird.service"}}, Include: "../base/base.yml"})
	if err != nil {
		t.Fatal(err)
	}
	systemd, ok := b.ignition["systemd"].(map[string]interface{})
	if !ok {
		t.Fatal("failed to construct ignition map")
	}

	actual := len(systemd["units"].([]interface{}))
	if actual != 2 {
		t.Errorf("in base.yml, defined a systemd unit, so expected length:%d, actual %d", 2, actual)
	}
}

func TestSabactlIgnitionTemplate(t *testing.T) {
	t.Run("constructIgnitionYAML", testIgnitionBuilderConstructIgnitionYAML)
	t.Run("constructFiles", testIgnitionBuilderConstructFile)
	t.Run("constructSystemd", testIgnitionBuilderConstructSystemd)
	t.Run("constructNetworkd", testIgnitionBuilderConstructNetworkd)
	t.Run("constructPasswd", testIgnitionBuilderConstructPasswd)
	t.Run("constructInclude", testIgnitionBuilderConstructInclude)
}
