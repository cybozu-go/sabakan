package sabakan

import (
	"context"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
)

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
