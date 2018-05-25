package etcd

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
)

const (
	etcdClientURL = "http://localhost:12379"
	etcdPeerURL   = "http://localhost:12380"
)

func testMain(m *testing.M) int {
	circleci := os.Getenv("CIRCLECI") == "true"
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

func newEtcdClient() (*clientv3.Client, error) {
	var clientURL string
	circleci := os.Getenv("CIRCLECI") == "true"
	if circleci {
		clientURL = "http://localhost:2379"
	} else {
		clientURL = etcdClientURL
	}
	return clientv3.New(clientv3.Config{
		Endpoints:   []string{clientURL},
		DialTimeout: 2 * time.Second,
	})
}

func testNewDriver(t *testing.T) (*driver, <-chan struct{}) {
	client, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	d := &driver{
		client: client,
		prefix: t.Name(),
		mi:     newMachinesIndex(),
	}
	ch := make(chan struct{}, 8) // buffers post-modify-done signals, up to 8
	go d.startWatching(context.Background(), ch)
	<-ch
	return d, ch
}
