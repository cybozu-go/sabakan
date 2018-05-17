// Package etcd implements sabakan model on etcd.
package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/sabakan"
)

type driver struct {
	client  *clientv3.Client
	watcher *clientv3.Client
	prefix  string
	mi      *machinesIndex
}

// NewModel returns sabakan.Model
func NewModel(client, watcher *clientv3.Client, prefix string) sabakan.Model {
	d := &driver{client, watcher, prefix, newMachinesIndex()}
	return sabakan.Model{
		Runner:  d,
		Storage: d,
		Machine: d,
		IPAM:    ipamDriver{d},
	}
}

// Run starts etcd watcher.  This should be called as a goroutine.
func (d *driver) Run(ctx context.Context) error {
	return d.startWatching(ctx)
}
