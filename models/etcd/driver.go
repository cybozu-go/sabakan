// Package etcd implements sabakan model on etcd.
package etcd

import (
	"context"
	"sync/atomic"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/sabakan"
)

type driver struct {
	client     *clientv3.Client
	watcher    *clientv3.Client
	prefix     string
	mi         *machinesIndex
	ipamConfig atomic.Value
	dhcpConfig atomic.Value
}

// NewModel returns sabakan.Model
func NewModel(client, watcher *clientv3.Client, prefix string) sabakan.Model {
	d := &driver{
		client:  client,
		watcher: watcher,
		prefix:  prefix,
		mi:      newMachinesIndex(),
	}
	return sabakan.Model{
		Runner:  d,
		Storage: d,
		Machine: d,
		IPAM:    ipamDriver{d},
		DHCP:    dhcpDriver{d},
	}
}

// Run starts etcd watcher.  This should be called as a goroutine.
func (d *driver) Run(ctx context.Context, ch chan<- struct{}) error {
	return d.startWatching(ctx, ch)
}
