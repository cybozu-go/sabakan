package e2e

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/cybozu-go/cmd"
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
		stopEtcd()
		log.Fatal(err)
	}
	defer func() {
		stopSabakan()
	}()

	time.Sleep(1 * time.Second)

	return m.Run()
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
	env := cmd.NewEnvironment(context.Background())
	env.Go(func(ctx context.Context) error {
		command := cmd.CommandContext(ctx, "../sabactl", args...)
		command.Stdout = bufio.NewWriter(&stdout)
		command.Stderr = bufio.NewWriter(&stderr)
		return command.Run()
	})
	env.Stop()
	err := env.Wait()
	return &stdout, &stderr, err
}
