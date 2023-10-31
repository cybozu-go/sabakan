package etcd

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/cybozu-go/etcdutil"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/v2"
	"github.com/cybozu-go/well"
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

func testNewDriver(t *testing.T) (*driver, <-chan struct{}) {
	client, err := newEtcdClient(t.Name() + "/")
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse("http://localhost:10080")
	if err != nil {
		t.Fatal(err)
	}
	d := &driver{
		client: client,
		httpclient: &well.HTTPClient{
			Client: &http.Client{},
		},
		mi:           newMachinesIndex(),
		advertiseURL: u,
	}
	ch := make(chan struct{}, 8) // buffers post-modify-done signals, up to 8
	go d.startStatelessWatcher(context.Background(), ch, nil)
	<-ch
	return d, ch
}

func initializeTestData(d *driver, ch <-chan struct{}) ([]*sabakan.Machine, error) {
	ctx := context.Background()
	config := &testIPAMConfig
	err := d.putIPAMConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	<-ch

	machines := []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "12345678",
			Labels: map[string]string{"product": "R630"},
			Role:   "worker",
		}),
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial:     "12345679",
			Labels:     map[string]string{"product": "R630"},
			Role:       "worker",
			RetireDate: time.Date(2018, time.November, 22, 1, 2, 3, 0, time.UTC),
		}),
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "123456789",
			Labels: map[string]string{"product": "R730"},
			Role:   "worker",
		}),
	}
	err = d.machineRegister(ctx, machines)
	if err != nil {
		return nil, err
	}
	<-ch
	<-ch
	return machines, nil
}
