package etcd

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/sabakan"
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

func (d *driver) startWatching(ctx context.Context, ch, indexCh chan<- struct{}) error {
	// obtain the current revision to avoid missing events.
	resp, err := d.client.Get(ctx, "/")
	if err != nil {
		return err
	}
	rev := resp.Header.Revision

	err = d.initIPAMConfig(ctx)
	if err != nil {
		return err
	}

	err = d.initDHCPConfig(ctx)
	if err != nil {
		return err
	}

	err = d.mi.init(ctx, d.client)
	if err != nil {
		return err
	}

	// notify the caller of the readiness
	ch <- struct{}{}

	rch := d.client.Watch(ctx, "/",
		clientv3.WithPrefix(),
		clientv3.WithPrevKV(),
		clientv3.WithRev(rev),
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
