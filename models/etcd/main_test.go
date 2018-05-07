package etcd

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
)

const etcdClientUrl = "http://localhost:12379"

func TestMain(m *testing.M) {
	etcdPath, err := ioutil.TempDir("", "sabakan-test")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("etcd",
		"--data-dir", etcdPath,
		"--listen-client-urls", etcdClientUrl,
		"--advertise-client-urls", etcdClientUrl)
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	code := m.Run()
	cmd.Process.Kill()
	cmd.Wait()
	os.RemoveAll(etcdPath)
	os.Exit(code)
}

func newEtcdClient() (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdClientUrl},
		DialTimeout: 2 * time.Second,
	})
}

func testNewDriver(t *testing.T) *Driver {
	client, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	watcher, err := newEtcdClient()
	if err != nil {
		t.Fatal(err)
	}
	return NewDriver(client, watcher, t.Name())
}
