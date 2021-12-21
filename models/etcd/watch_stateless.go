package etcd

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cybozu-go/sabakan/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func (d *driver) initIPAMConfig(ctx context.Context) error {
	resp, err := d.client.Get(ctx, KeyIPAM)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return nil
	}

	config := new(sabakan.IPAMConfig)
	err = json.Unmarshal(resp.Kvs[0].Value, config)
	if err != nil {
		return err
	}

	d.ipamConfig.Store(config)
	return nil
}

func (d *driver) initDHCPConfig(ctx context.Context) error {
	resp, err := d.client.Get(ctx, KeyDHCP)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return nil
	}

	config := new(sabakan.DHCPConfig)
	err = json.Unmarshal(resp.Kvs[0].Value, config)
	if err != nil {
		return err
	}

	d.dhcpConfig.Store(config)
	return nil
}

func (d *driver) initStateless(ctx context.Context, ch chan<- struct{}) (int64, error) {
	defer func() {
		// notify the caller of the readiness
		ch <- struct{}{}
	}()

	// obtain the current revision to avoid missing events.
	resp, err := d.client.Get(ctx, "/")
	if err != nil {
		return 0, err
	}
	rev := resp.Header.Revision

	err = d.initIPAMConfig(ctx)
	if err != nil {
		return 0, err
	}

	err = d.initDHCPConfig(ctx)
	if err != nil {
		return 0, err
	}

	err = d.mi.init(ctx, d.client)
	if err != nil {
		return 0, err
	}

	return rev, nil
}

// startStatelessWatcher is a goroutine to begin watching for etcd events.
//
// This goroutine does not keep states between restarts; i.e. it loads
// the up-to-date database entries in memory and keeps updating them.
func (d *driver) startStatelessWatcher(ctx context.Context, ch, indexCh chan<- struct{}) error {
	rev, err := d.initStateless(ctx, ch)
	if err != nil {
		return err
	}

	rch := d.client.Watch(ctx, "",
		clientv3.WithPrefix(),
		clientv3.WithPrevKV(),
		clientv3.WithRev(rev+1),
	)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			var err error
			key := string(ev.Kv.Key)
			switch {
			case strings.HasPrefix(key, KeyMachines):
				err = d.handleMachines(ev)
			case key == KeyDHCP:
				err = d.handleDHCPConfig(ev)
			case key == KeyIPAM:
				err = d.handleIPAMConfig(ev)
			case strings.HasPrefix(key, KeyImages):
				select {
				case indexCh <- struct{}{}:
				default:
				}
			}
			if err != nil {
				panic(err)
			}
		}
		select {
		case ch <- struct{}{}:
		default:
		}
	}

	return nil
}
