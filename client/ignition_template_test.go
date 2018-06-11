package client

import (
	"testing"
)

func TestIgnitionBuilderConstructIgnitionYAML(t *testing.T) {
	b := ignitionBuilder{baseDir: "../testdata/ignitions", ignition: make(map[string]interface{})}

	tests := []struct {
		name    string
		source  *ignitionSource
		wantErr bool
	}{
		{name: "passwd", source: &ignitionSource{Passwd: "passwd.yml"}, wantErr: false},
		{name: "files", source: &ignitionSource{Files: []string{"/etc/hostname"}}, wantErr: false},
		{name: "systemd", source: &ignitionSource{Systemd: []systemd{{Source: "bird.service"}}}, wantErr: false},
		{name: "networkd", source: &ignitionSource{Networkd: []string{"10-node0.netdev"}}, wantErr: false},
		{name: "include", source: &ignitionSource{Include: "base.yml"}, wantErr: false},

		{name: "passwd not found", source: &ignitionSource{Passwd: "nonexists_file.yml"}, wantErr: true},
		{name: "files not found", source: &ignitionSource{Files: []string{"/etc/not_file"}}, wantErr: true},
		{name: "systemd not found", source: &ignitionSource{Systemd: []systemd{{Source: "hoge.service"}}}, wantErr: true},
		{name: "networkd not found", source: &ignitionSource{Networkd: []string{"nonexists.netdev"}}, wantErr: true},
		{name: "include not found", source: &ignitionSource{Include: "nonexits_base.yml"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := b.constructIgnitionYAML(tt.source); (err != nil) != tt.wantErr {
				t.Errorf("ignitionBuilder.constructIgnitionYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIgnitionBuilderConstructFile(t *testing.T) {
	b := ignitionBuilder{baseDir: "../testdata/ignitions", ignition: make(map[string]interface{})}
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
		t.Errorf("constructFiles appends a unit, so expected length:%d, actual %d", 2, actual)
	}
}

func TestIgnitionBuilderConstructSystemd(t *testing.T) {
	b := ignitionBuilder{baseDir: "../testdata/ignitions", ignition: make(map[string]interface{})}
	s := systemd{true, "bird.service"}
	err := b.constructSystemd(s)
	if err != nil {
		t.Fatal(err)
	}
	err = b.constructSystemd(s)
	if err != nil {
		t.Fatal(err)
	}
	systemd, ok := b.ignition["systemd"].(map[string]interface{})
	if !ok {
		t.Fatal("failed to construct ignition map")
	}

	actual := len(systemd["units"].([]interface{}))
	if actual != 2 {
		t.Errorf("constructSystemd appends a unit, so expected length:%d, actual %d", 2, actual)
	}
}

func TestIgnitionBuilderConstructNetworkd(t *testing.T) {
	b := ignitionBuilder{baseDir: "../testdata/ignitions", ignition: make(map[string]interface{})}
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
		t.Errorf("constructNetworkd appends a unit, so expected length:%d, actual %d", 2, actual)
	}
}

func TestIgnitionBuilderConstructPasswd(t *testing.T) {
	b := ignitionBuilder{baseDir: "../testdata/ignitions", ignition: make(map[string]interface{})}
	src := "passwd.yml"
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

func TestIgnitionBuilderConstructInclude(t *testing.T) {
	b := ignitionBuilder{baseDir: "../testdata/ignitions", ignition: make(map[string]interface{})}

	err := b.constructIgnitionYAML(&ignitionSource{Systemd: []systemd{{Source: "bird.service"}}, Include: "base.yml"})
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
