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

	etcdPath, err := ioutil.TempDir("", "sabakan-test")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("etcd",
		"--data-dir", etcdPath,
		"--initial-cluster", "default="+etcdPeerURL,
		"--listen-peer-urls", etcdPeerURL,
		"--initial-advertise-peer-urls", etcdPeerURL,
		"--listen-client-urls", etcdClientURL,
		"--advertise-client-urls", etcdClientURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
		os.RemoveAll(etcdPath)
	}()

	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func runSabakan() {
	servers := etcdClientURL
	if circleci {
		servers = "http://localhost:2379"
	}
	cmd.Go(func(ctx context.Context) error {
		command := cmd.CommandContext(ctx,
			"go", "run", "../cmd/sabakan/main.go",
			"-dhcp-interface", "lo", "-dhcp-bind", "0.0.0.0:10067",
			"-etcd-servers", servers,
		)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		return command.Run()
	})

	time.Sleep(1 * time.Second)
}

func runSabactl(args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	var stdout, stderr bytes.Buffer
	env := cmd.NewEnvironment(context.Background())
	env.Go(func(ctx context.Context) error {
		params := []string{
			"run",
			"../cmd/sabactl/main.go",
			"../cmd/sabactl/machines.go",
			"../cmd/sabactl/remote_config.go",
		}
		params = append(params, args...)
		command := cmd.CommandContext(ctx, "go", params...)
		command.Stdout = bufio.NewWriter(&stdout)
		command.Stderr = bufio.NewWriter(&stderr)
		return command.Run()
	})
	env.Stop()
	err := env.Wait()
	return &stdout, &stderr, err
}
