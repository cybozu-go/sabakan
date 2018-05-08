// Package etcd implements sabakan model on etcd.
package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
)

// Driver implements sabakan model.
type Driver struct {
	client  *clientv3.Client
	watcher *clientv3.Client
	prefix  string
	mi      *machinesIndex
}

// NewDriver returns an object that implements all interfaces required for sabakan.Model.
func NewDriver(client, watcher *clientv3.Client, prefix string) *Driver {
	return &Driver{client, watcher, prefix, newMachinesIndex()}
}

// Run starts etcd watcher.  This should be called as a goroutine.
func (d *Driver) Run(ctx context.Context) error {
	return d.startWatching(ctx)
}
