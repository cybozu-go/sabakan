package etcd

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/etcdutil"
	"github.com/cybozu-go/sabakan"
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

func newEtcdClient(prefix string) (*clientv3.Client, error) {
	var clientURL string
	circleci := os.Getenv("CIRCLECI") == "true"
	if circleci {
		clientURL = "http://localhost:2379"
	} else {
		clientURL = etcdClientURL
	}

	cfg := etcdutil.NewConfig(prefix)
	cfg.Endpoints = []string{clientURL}
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
		httpclient: &cmd.HTTPClient{
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
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345678", Labels: map[string]string{"product": "R630"}, Role: "worker"}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345679", Labels: map[string]string{"product": "R630"}, Role: "worker"}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "123456789", Labels: map[string]string{"product": "R730"}, Role: "worker"}),
	}
	err = d.machineRegister(ctx, machines)
	if err != nil {
		return nil, err
	}
	<-ch
	<-ch
	return machines, nil
}
