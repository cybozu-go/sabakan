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
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	time.Sleep(100 * time.Millisecond)

	stdout, stderr, err = runSabactl("dhcp", "get")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
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
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
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
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	time.Sleep(100 * time.Millisecond)

	stdout, stderr, err = runSabactl("ipam", "get")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
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
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
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
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	time.Sleep(100 * time.Millisecond)

	specs := []*sabakan.MachineSpec{
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
	stdout, stderr, err = runSabactlWithFile(t, specs, "machines", "create")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	stdout, stderr, err = runSabactl("machines", "get", "--serial", "12345678")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}
	var gotMachines []sabakan.Machine
	err = json.Unmarshal(stdout.Bytes(), &gotMachines)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotMachines) != 1 {
		t.Fatal("machine not found")
	}

	stdout, stderr, err = runSabactl("machines", "get-state", "12345678")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}
	gotState := sabakan.MachineState(stdout.String())
	if err != nil {
		t.Fatal(err)
	}
	if gotState != sabakan.StateHealthy {
		t.Fatal("unexpected machine state: ", gotState)
	}

	stdout, stderr, err = runSabactl("machines", "remove", "12345678")
	code = exitCode(err)
	if code != client.ExitResponse5xx {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	stdout, stderr, err = runSabactl("machines", "set-state", "12345678", "retiring")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	stdout, stderr, err = runSabactl("crypts", "delete", "-force", "12345678")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	stdout, stderr, err = runSabactl("machines", "remove", "12345678")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("exit code:", code)
	}

	stdout, stderr, err = runSabactl("machines", "get", "--serial", "12345678")
	code = exitCode(err)
	if code != client.ExitNotFound {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("machine not removed", code)
	}
}

func testSabactlImages(t *testing.T) {
	// upload image
	kernelFile, err := ioutil.TempFile("", "sabakan-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(kernelFile.Name())
	kernelFile.Close()

	initrdFile, err := ioutil.TempFile("", "sabakan-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(initrdFile.Name())
	initrdFile.Close()

	stdout, stderr, err := runSabactl("images", "upload", "1234.1.0", kernelFile.Name(), initrdFile.Name())
	code := exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to upload image", code)
	}

	// retrieve index
	stdout, stderr, err = runSabactl("images", "index")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to get index of images", code)
	}

	var got sabakan.ImageIndex
	err = json.Unmarshal(stdout.Bytes(), &got)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatal("length of index != 1", len(got))
	}
	if got[0].ID != "1234.1.0" {
		t.Error("wrong ID in index", got[0].ID)
	}
	if len(got[0].URLs) != 1 {
		t.Error("wrong number of URLs in index", len(got[0].URLs))
	}
	if !got[0].Exists {
		t.Error("index says that local image is not stored")
	}

	// delete image
	stdout, stderr, err = runSabactl("images", "delete", "1234.1.0")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to delete image", code)
	}

	// retrieve (empty) index
	stdout, stderr, err = runSabactl("images", "index")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to get index of images", code)
	}

	err = json.Unmarshal(stdout.Bytes(), &got)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Error("index was not empty")
	}
}

func testSabactlAssets(t *testing.T) {
	// upload asset
	file, err := ioutil.TempFile("", "sabakan-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString("bar")
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	stdout, stderr, err := runSabactl("assets", "upload", "foo", file.Name())
	code := exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to upload asset", code)
	}

	// retrieve index
	stdout, stderr, err = runSabactl("assets", "index")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to get index of assets", code)
	}

	var index []string
	err = json.Unmarshal(stdout.Bytes(), &index)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(index, []string{"foo"}) {
		t.Error("wrong index", index)
	}

	// retrieve asset info
	stdout, stderr, err = runSabactl("assets", "info", "foo")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to get meta data of asset", code)
	}

	var asset sabakan.Asset
	err = json.Unmarshal(stdout.Bytes(), &asset)
	if err != nil {
		t.Fatal(err)
	}
	if asset.Name != "foo" {
		t.Error("asset.Name != foo:", asset.Name)
	}
	if asset.ID == 0 {
		t.Error("asset.ID should not be 0")
	}
	if asset.ContentType != "application/octet-stream" {
		t.Error("asset.ContentType != application/octet-stream:", asset.ContentType)
	}
	if asset.Sha256 != "fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9" {
		t.Error("wrong Sha256:", asset.Sha256)
	}
	if len(asset.URLs) != 1 {
		t.Error("wrong number of URLs:", asset.URLs)
	}
	if !asset.Exists {
		t.Error("asset info says that local asset file is not stored")
	}

	// delete asset
	stdout, stderr, err = runSabactl("assets", "delete", "foo")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to delete asset", code)
	}

	// retrieve empty index
	stdout, stderr, err = runSabactl("assets", "index")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to get index of assets", code)
	}

	err = json.Unmarshal(stdout.Bytes(), &index)
	if err != nil {
		t.Fatal(err)
	}
	if len(index) != 0 {
		t.Error("index was not empty")
	}
}

func testSabactlIgnitions(t *testing.T) {
	saved := `ignition:
  version: 2.2.0
`
	stdout, stderr, err := runSabactl("ignitions", "set", "-f", "../testdata/test/empty.yml", "cs")
	code := exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to set ignition template", code)
	}
	stdout, stderr, err = runSabactl("ignitions", "set", "-f", "../testdata/test/test.yml", "cs")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to set ignition template", code)
	}

	stdout, stderr, err = runSabactl("ignitions", "get", "cs")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to get ignition template IDs of cs", code)
	}
	var ids []string
	err = json.NewDecoder(stdout).Decode(&ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Error("expected:1, actual:", len(ids))
	}

	stdout, stderr, err = runSabactl("ignitions", "cat", "cs", ids[0])
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to cat ignition template", code)
	}
	if stdout.String() != saved {
		t.Error("stdout.String() != saved", stdout.String())
	}

	stdout, stderr, err = runSabactl("ignitions", "delete", "cs", ids[0])
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to delete an ignition template", code)
	}

	stdout, stderr, err = runSabactl("ignitions", "get", "cs")
	code = exitCode(err)
	if code != client.ExitSuccess {
		t.Log("stdout:", stdout.String())
		t.Log("stderr:", stderr.String())
		t.Fatal("failed to get ignition template IDs of cs", code)
	}
	ids = []string{}
	err = json.NewDecoder(stdout).Decode(&ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Error("expected:1, actual:", len(ids))
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
	t.Run("Images", testSabactlImages)
	t.Run("Assets", testSabactlAssets)
	t.Run("Ignitions", testSabactlIgnitions)
}
