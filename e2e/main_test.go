package e2e

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/cybozu-go/sabakan/client"
)

const (
	etcdClientURL = "http://localhost:12379"
	etcdPeerURL   = "http://localhost:12380"
)

var circleci = false

func init() {
	circleci = os.Getenv("CIRCLECI") == "true"
}

func testMain(m *testing.M) int {
	if circleci {
		code := m.Run()
		os.Exit(code)
	}

	stopEtcd := runEtcd()
	defer func() {
		stopEtcd()
	}()

	stopSabakan, err := runSabakan()
	if err != nil {
		// log.Fatal() uses os.Exit(), and it does not process defer.
		stopEtcd()
		log.Fatal(err)
	}
	defer func() {
		stopSabakan()
	}()

	// wait for sabakan
	for i := 0; i < 10; i++ {
		_, _, err = runSabactl("remote-config", "get")
		code := exitCode(err)
		if code == client.ExitNotFound {
			return m.Run()
		}
		time.Sleep(1 * time.Second)
	}

	stopEtcd()
	stopSabakan()
	log.Fatal(err)
	panic("unreachable")
}

func runEtcd() func() {
	etcdPath, err := ioutil.TempDir("", "sabakan-test")
	if err != nil {
		log.Fatal(err)
	}
	command := exec.Command("etcd",
		"--data-dir", etcdPath,
		"--initial-cluster", "default="+etcdPeerURL,
		"--listen-peer-urls", etcdPeerURL,
		"--initial-advertise-peer-urls", etcdPeerURL,
		"--listen-client-urls", etcdClientURL,
		"--advertise-client-urls", etcdClientURL)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err = command.Start()
	if err != nil {
		log.Fatal(err)
	}

	return func() {
		command.Process.Kill()
		command.Wait()
		os.RemoveAll(etcdPath)
	}
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func runSabakan() (func(), error) {
	servers := etcdClientURL
	if circleci {
		servers = "http://localhost:2379"
	}
	command := exec.Command("../sabakan",
		"-dhcp-interface", "lo", "-dhcp-bind", "0.0.0.0:10067",
		"-etcd-servers", servers,
	)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Start()
	if err != nil {
		return nil, err
	}

	return func() {
		command.Process.Kill()
		command.Wait()
	}, nil
}

func runSabactl(args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer
	command := exec.Command("../sabactl", args...)
	command.Stdout = bufio.NewWriter(&stdout)
	command.Stderr = bufio.NewWriter(&stderr)
	return &stdout, &stderr, command.Run()
}
