package etcd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/sabakan"
)

const (
	etcdClientURL = "https://localhost:12379"
	etcdPeerURL   = "https://localhost:12380"
	etcdCA        = "../../testdata/certs/ca.crt"
	etcdCert      = "../../testdata/certs/server.crt"
	etcdKey       = "../../testdata/certs/server.key.insecure"
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
		"--client-cert-auth",
		"--trusted-ca-file", etcdCA,
		"--cert-file", etcdCert,
		"--key-file", etcdKey,
		"--peer-trusted-ca-file", etcdCA,
		"--peer-cert-file", etcdCert,
		"--peer-key-file", etcdKey,
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
		clientURL = "https://localhost:2379"
	} else {
		clientURL = etcdClientURL
	}
	cert, err := tls.LoadX509KeyPair(etcdCert, etcdKey)
	if err != nil {
		return nil, err
	}
	rootCACert, err := ioutil.ReadFile(etcdCA)
	if err != nil {
		return nil, err
	}
	rootCAs := x509.NewCertPool()
	ok := rootCAs.AppendCertsFromPEM(rootCACert)
	if !ok {
		return nil, errors.New("Failed to parse PEM file")
	}
	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootCAs,
	}
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{clientURL},
		DialTimeout: 2 * time.Second,
		TLS:         tlsCfg,
	})
	if err != nil {
		return nil, err
	}
	c.KV = namespace.NewKV(c.KV, prefix)
	c.Watcher = namespace.NewWatcher(c.Watcher, prefix)
	c.Lease = namespace.NewLease(c.Lease, prefix)
	return c, nil
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
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345678", Product: "R630", Role: "worker"}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345679", Product: "R630", Role: "worker"}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "123456789", Product: "R730", Role: "worker"}),
	}
	err = d.machineRegister(ctx, machines)
	if err != nil {
		return nil, err
	}
	<-ch
	<-ch
	return machines, nil
}
