// Package etcd implements sabakan model on etcd.
package etcd

import (
	"context"
	"net/url"
	"path"
	"sync/atomic"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/sabakan"
)

type driver struct {
	client       *clientv3.Client
	dataDir      string
	advertiseURL *url.URL
	mi           *machinesIndex
	ipamConfig   atomic.Value
	dhcpConfig   atomic.Value
}

// NewModel returns sabakan.Model
func NewModel(client *clientv3.Client, dataDir string, advertiseURL *url.URL) sabakan.Model {
	d := &driver{
		client:       client,
		dataDir:      dataDir,
		advertiseURL: advertiseURL,
		mi:           newMachinesIndex(),
	}
	return sabakan.Model{
		Runner:   d,
		Storage:  d,
		Machine:  d,
		IPAM:     ipamDriver{d},
		DHCP:     dhcpDriver{d},
		Image:    imageDriver{d},
		Asset:    assetDriver{d},
		Ignition: d,
	}
}

func (d *driver) myURL(p ...string) string {
	u := *d.advertiseURL
	u.Path = path.Join(p...)
	return u.String()
}

// Run starts etcd watcher.  This should be called as a goroutine.
//
// The watcher sends an object when it completes the initialization
// and starts watching.  Callers must receive the object.
//
// The watcher continue to send an event every time it handles an event
// if and only if sending to ch is not going to be blocked.
// This can be used by tests to synchronize with the watcher.
func (d *driver) Run(ctx context.Context, ch chan<- struct{}) error {
	imageIndexCh := make(chan struct{}, 1)

	env := cmd.NewEnvironment(ctx)
	env.Go(func(ctx context.Context) error {
		return d.startWatching(ctx, ch, imageIndexCh)
	})
	env.Go(func(ctx context.Context) error {
		return d.startImageUpdater(ctx, imageIndexCh)
	})
	env.Stop()

	return env.Wait()
}
