// Package etcd implements sabakan model on etcd.
package etcd

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"sync/atomic"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/sabakan"
)

type driver struct {
	client       *clientv3.Client
	httpclient   *cmd.HTTPClient
	dataDir      string
	advertiseURL *url.URL
	mi           *machinesIndex
	ipamConfig   atomic.Value
	dhcpConfig   atomic.Value
}

// NewModel returns sabakan.Model
func NewModel(client *clientv3.Client, dataDir string, advertiseURL *url.URL) sabakan.Model {
	d := &driver{
		client: client,
		httpclient: &cmd.HTTPClient{
			Client: &http.Client{},
		},
		dataDir:      dataDir,
		advertiseURL: advertiseURL,
		mi:           newMachinesIndex(),
	}
	return sabakan.Model{
		Runner:   d,
		Storage:  d,
		Machine:  machineDriver{d},
		IPAM:     ipamDriver{d},
		DHCP:     dhcpDriver{d},
		Image:    imageDriver{d},
		Asset:    assetDriver{d},
		Log:      logDriver{d},
		Ignition: d,
	}
}

func (d *driver) myURL(p ...string) string {
	u := *d.advertiseURL
	u.Path = path.Join(p...)
	return u.String()
}

func (d *driver) pullURL(ctx context.Context, u string) (*http.Response, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	return d.httpclient.Do(req)
}

// EventPool is a pool of events.
type EventPool struct {
	Rev    int64
	Events []*clientv3.Event
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
	epCh := make(chan EventPool)

	env := cmd.NewEnvironment(ctx)

	// stateless watcher and its consumer
	env.Go(func(ctx context.Context) error {
		return d.startStatelessWatcher(ctx, ch, imageIndexCh)
	})
	env.Go(func(ctx context.Context) error {
		return d.startImageUpdater(ctx, imageIndexCh)
	})

	// stateful watcher and its consumer
	env.Go(func(ctx context.Context) error {
		return d.startStatefulWatcher(ctx, epCh)
	})
	env.Go(func(ctx context.Context) error {
		return d.startAssetUpdater(ctx, epCh)
	})

	// log compaction
	env.Go(d.logCompactor)

	env.Stop()

	return env.Wait()
}
