package e2e

import (
	"bytes"
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

func runSabactlWithFile(t *testing.T, data interface{}, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	f, err := ioutil.TempFile("", "sabakan-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	err = json.NewEncoder(f).Encode(data)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	args = append(args, "-f", f.Name())

	return runSabactl(args...)
}

func testSabactlRemoteConfig(t *testing.T) {
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
	stdout, stderr, err := runSabactlWithFile(t, &conf, "remote-config", "set")
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

	var badConf = sabakan.IPAMConfig{
		NodeIPv4Offset:  "10.69.0.0/26",
		NodeRackShift:   6,
		NodeIndexOffset: 3,
		BMCIPv4Offset:   "10.72.17.0/27",
		BMCRackShift:    5,
		NodeIPPerNode:   3,
		BMCIPPerNode:    1,
	}
	stdout, stderr, err = runSabactlWithFile(t, &badConf, "remote-config", "set")
	code = exitCode(err)
	if code != client.ExitResponse4xx {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}
}

func testSabactlMachines(t *testing.T) {
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
	stdout, stderr, err := runSabactlWithFile(t, &conf, "remote-config", "set")
	code := exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	machines := []sabakan.Machine{
		{
			Serial:     "12345678",
			Product:    "R730xd",
			Datacenter: "ty3",
			Role:       "worker",
		},
		//{
		//	Serial:     "abcdefg",
		//	Product:    "R730xd",
		//	Datacenter: "ty3",
		//	Role:       "boot",
		//},
	}
	stdout, stderr, err = runSabactlWithFile(t, &machines, "machines", "create")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	stdout, stderr, err = runSabactl("machines", "get", "--serial", "12345678")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}
	var gotMachines []sabakan.Machine
	json.Unmarshal(stdout.Bytes(), &gotMachines)
	if len(gotMachines) != 1 {
		t.Fatal("machine not found")
	}

	stdout, stderr, err = runSabactl("machines", "delete", "12345678")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	stdout, stderr, err = runSabactl("machines", "get", "--serial", "12345678")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}
	json.Unmarshal(stdout.Bytes(), &gotMachines)
	if len(gotMachines) != 0 {
		t.Fatal("machine not deleted")
	}
}

func TestSabactl(t *testing.T) {
	t.Run("sabactl remote-config", testSabactlRemoteConfig)
	t.Run("sabactl machines", testSabactlMachines)
}
