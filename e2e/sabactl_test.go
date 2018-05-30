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
	"time"

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

func testSabactlDHCP(t *testing.T) {
	var conf = sabakan.DHCPConfig{
		GatewayOffset: 100,
	}
	stdout, stderr, err := runSabactlWithFile(t, &conf, "dhcp", "set")
	code := exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	time.Sleep(100 * time.Millisecond)

	stdout, stderr, err = runSabactl("dhcp", "get")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	var got sabakan.DHCPConfig
	err = json.Unmarshal(stdout.Bytes(), &got)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(conf, got) {
		t.Error("unexpected config", got)
	}

	var badConf = sabakan.DHCPConfig{}
	stdout, stderr, err = runSabactlWithFile(t, &badConf, "dhcp", "set")
	code = exitCode(err)
	if code != client.ExitResponse4xx {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}
}

func testSabactlIPAM(t *testing.T) {
	var conf = sabakan.IPAMConfig{
		MaxNodesInRack:  28,
		NodeIPv4Pool:    "10.69.0.0/20",
		NodeRangeSize:   6,
		NodeRangeMask:   26,
		NodeIndexOffset: 3,
		NodeIPPerNode:   3,
		BMCIPv4Pool:     "10.72.16.0/20",
		BMCRangeSize:    5,
		BMCRangeMask:    20,
	}
	stdout, stderr, err := runSabactlWithFile(t, &conf, "ipam", "set")
	code := exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	time.Sleep(100 * time.Millisecond)

	stdout, stderr, err = runSabactl("ipam", "get")
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
		MaxNodesInRack:  0,
		NodeIPv4Pool:    "10.69.0.0/20",
		NodeRangeSize:   6,
		NodeRangeMask:   26,
		NodeIndexOffset: 3,
		NodeIPPerNode:   3,
		BMCIPv4Pool:     "10.72.16.0/20",
		BMCRangeSize:    5,
		BMCRangeMask:    20,
	}
	stdout, stderr, err = runSabactlWithFile(t, &badConf, "ipam", "set")
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
		NodeIPv4Pool:    "10.69.0.0/20",
		NodeRangeSize:   6,
		NodeRangeMask:   26,
		NodeIndexOffset: 3,
		NodeIPPerNode:   3,
		BMCIPv4Pool:     "10.72.16.0/20",
		BMCRangeSize:    5,
		BMCRangeMask:    20,
	}
	stdout, stderr, err := runSabactlWithFile(t, &conf, "ipam", "set")
	code := exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	time.Sleep(100 * time.Millisecond)

	machines := []sabakan.Machine{
		{
			Serial:     "12345678",
			Product:    "R730xd",
			Datacenter: "ty3",
			Role:       "worker",
			BMC: sabakan.MachineBMC{
				Type: sabakan.BmcIdrac9,
			},
		},
		{
			Serial:     "abcdefg",
			Product:    "R730xd",
			Datacenter: "ty3",
			Role:       "boot",
			BMC: sabakan.MachineBMC{
				Type: sabakan.BmcIpmi2,
			},
		},
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

	stdout, stderr, err = runSabactl("machines", "remove", "12345678")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	stdout, stderr, err = runSabactl("machines", "get", "--serial", "12345678")
	code = exitCode(err)
	if code != client.ExitNotFound {
		t.Error("stdout:", stdout.String())
		t.Error("stderr:", stderr.String())
		t.Fatal("machine not removed", code)
	}
}

func TestSabactl(t *testing.T) {
	_, err := os.Stat("../sabactl")
	if err != nil {
		t.Skip("sabactl executable not found")
	}
	t.Run("DHCP", testSabactlDHCP)
	t.Run("IPAM", testSabactlIPAM)
	t.Run("Machines", testSabactlMachines)
}
