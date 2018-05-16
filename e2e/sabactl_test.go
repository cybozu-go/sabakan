package e2e

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func testSabactlRemoteConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "sabakan-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	var conf = sabakan.IPAMConfig{
		NodeIPv4Offset: "10.0.0.0/24",
		NodeRackShift:  4,
		NodeIPPerNode:  3,
		BMCIPv4Offset:  "10.10.0.0/24",
		BMCRackShift:   2,
		BMCIPPerNode:   1,
	}
	err = json.NewEncoder(f).Encode(conf)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	stdout, stderr, err := runSabactl("remote-config", "set", "-f", f.Name())
	if err != nil {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal(err)
	}

	stdout, stderr, err = runSabactl("remote-config", "get")
	if err != nil {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal(err)
	}

	var got sabakan.IPAMConfig
	err = json.Unmarshal(stdout.Bytes(), &got)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, got) {
		t.Error("unexpected config", got)
	}
}

func TestSabactl(t *testing.T) {
	t.Run("sabactl remote-config", testSabactlRemoteConfig)
}
