package e2e

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

const (
	etcdClientURL = "http://localhost:12379"
	etcdPeerURL   = "http://localhost:12380"
)

var circleci = false

func init() {
	circleci = os.Getenv("CIRCLECI") == "true"
}

func testMain(m *testing.M) (int, error) {
	stopEtcd := runEtcd()
	defer func() {
		stopEtcd()
	}()

	stopSabakan, err := runSabakan()
	if err != nil {
		return 0, err
	}
	defer func() {
		stopSabakan()
	}()

	return m.Run(), nil
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
	if circleci {
		code := m.Run()
		os.Exit(code)
	}

	if len(os.Getenv("RUN_E2E")) == 0 {
		os.Exit(0)
	}

	status, err := testMain(m)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(status)
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

	// wait for startup
	for i := 0; i < 10; i++ {
		resp, err := http.Get("http://localhost:8888/api/v1/config/ipam")
		if err == nil {
			resp.Body.Close()
			return func() {
				command.Process.Kill()
				command.Wait()
			}, nil
		}
		time.Sleep(1 * time.Second)
	}

	return nil, err
}

func runSabactl(args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer
	command := exec.Command("../sabactl", args...)
	command.Stdout = bufio.NewWriter(&stdout)
	command.Stderr = bufio.NewWriter(&stderr)
	return &stdout, &stderr, command.Run()
}
