package sabakan

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var (
	flagEtcdServers = flag.String("etcd-servers", "http://localhost:2379", "URLs of the backend etcd")
	flagEtcdPrefix  = flag.String("etcd-prefix", "/sabakan-test", "etcd prefix")
)

func TestMain(m *testing.M) {
	err := setupEtcd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func setupEtcd() error {
	etcd, err := newEtcdClient()
	if err != nil {
		return err
	}
	defer etcd.Close()

	_, err = etcd.Delete(context.Background(), *flagEtcdPrefix, clientv3.WithPrefix())
	return err
}

func newEtcdClient() (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(*flagEtcdServers, ","),
		DialTimeout: 2 * time.Second,
	})
}
