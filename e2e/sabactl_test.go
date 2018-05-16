package e2e

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"syscall"
	"testing"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

func exitCode(err error) subcommands.ExitStatus {
	if err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				return subcommands.ExitStatus(s.ExitStatus())
			}
			// unexpected error; not Unix?
			panic(err)
		}
		// exec itself failed, e.g. command not found
		panic(err)
	}
	return client.ExitSuccess
}

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
	code := exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	stdout, stderr, err = runSabactl("remote-config", "get")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
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
