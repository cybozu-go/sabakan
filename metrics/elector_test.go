package metrics

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/cybozu-go/etcdutil"
	"github.com/cybozu-go/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	etcdClientURL = "http://localhost:12379"
	etcdPeerURL   = "http://localhost:12380"
)

func testMain(m *testing.M) int {
	etcdDataDir, err := os.MkdirTemp("", "sabakan-test")
	if err != nil {
		log.ErrorExit(err)
	}

	cmd := exec.Command("etcd",
		"--data-dir", etcdDataDir,
		"--initial-cluster", "default="+etcdPeerURL,
		"--listen-peer-urls", etcdPeerURL,
		"--initial-advertise-peer-urls", etcdPeerURL,
		"--listen-client-urls", etcdClientURL,
		"--advertise-client-urls", etcdClientURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.ErrorExit(err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
		os.RemoveAll(etcdDataDir)
	}()

	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func newEtcdClient(prefix string) (*clientv3.Client, error) {
	cfg := etcdutil.NewConfig(prefix)
	cfg.Endpoints = []string{etcdClientURL}
	return etcdutil.NewClient(cfg)
}

func TestEtcdLeaderElectionFailover(t *testing.T) {
	c, err := newEtcdClient(t.Name() + "/")
	if err != nil {
		t.Fatal(err)
	}

	e1 := NewLeaderElector(c, "/sabakan/test-leader", "node-1", 60*time.Second)
	e2 := NewLeaderElector(c, "/sabakan/test-leader", "node-2", 60*time.Second)
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel1()
		cancel2()
		e1.Close()
		e2.Close()
	})

	go func() {
		e1.Run(ctx1)
	}()
	go func() {
		e2.Run(ctx2)
	}()

	err = awaitTrue(5*time.Second, func() bool {
		return e1.IsLeader() || e2.IsLeader()
	})
	if err != nil {
		t.Fatal("Time out (5s) waiting for any elector to become leader")
	}

	if e1.IsLeader() {
		cancel1()
	} else {
		cancel2()
	}

	err = awaitTrue(5*time.Second, func() bool {
		return e1.IsLeader() || e2.IsLeader()
	})
	if err != nil {
		t.Fatal("After stoptting the initial leader, no new leader was elected within 5 seconds")
	}
}

func awaitTrue(d time.Duration, f func() bool) error {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if f() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting condition")
}
