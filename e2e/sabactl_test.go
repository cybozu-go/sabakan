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
		MaxNodesInRack:  28,
		NodeIPv4Offset:  "10.69.0.0/26",
		NodeRackShift:   6,
		NodeIndexOffset: 3,
		BMCIPv4Offset:   "10.72.17.0/27",
		BMCRackShift:    5,
		NodeIPPerNode:   3,
		BMCIPPerNode:    1,
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
