package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

var (
	errLostOwner = errors.New("lost ownership during conversion")
)

func (d *driver) convertTo2Machines(ctx context.Context, mu *concurrency.Mutex, ipam *sabakan.IPAMConfig) error {
	const limitMachines = 20
	key := KeyMachines
	endKey := clientv3.GetPrefixRangeEnd(KeyMachines)
	resp, err := d.client.Get(ctx, key,
		clientv3.WithRange(endKey),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		clientv3.WithLimit(limitMachines),
	)
	if err != nil {
		return err
	}
	rev := resp.Header.Revision //  to retrieve following pages at the same revision.
	kvs := resp.Kvs
	var ops []clientv3.Op

REDO:
	if len(kvs) == 0 {
		return nil
	}

	ops = make([]clientv3.Op, len(kvs))
	for i, kv := range kvs {
		var m sabakan.Machine
		err = json.Unmarshal(kv.Value, &m)
		if err != nil {
			return fmt.Errorf("failed to unmarshal %s: %v", string(kv.Key[len(KeyMachines):]), err)
		}

		// fill Machine.Info
		ipam.GenerateIP(&m)

		data, err := json.Marshal(m)
		if err != nil {
			return err
		}
		ops[i] = clientv3.OpPut(KeyMachines+m.Spec.Serial, string(data))
	}
	tresp, err := d.client.Txn(ctx).If(mu.IsOwner()).
		Then(ops...).
		Commit()
	if err != nil {
		return err
	}
	if !tresp.Succeeded {
		return errLostOwner
	}

	if resp.More {
		resp, err = d.client.Get(ctx, string(resp.Kvs[len(resp.Kvs)-1].Key),
			clientv3.WithRange(endKey),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
			clientv3.WithLimit(limitMachines),
			clientv3.WithRev(rev),
		)
		if err != nil {
			return err
		}

		// ignore the first key
		kvs = resp.Kvs[1:]
		goto REDO
	}

	return nil
}

func (d *driver) convertTo2(ctx context.Context, mu *concurrency.Mutex) error {
	// we cannot use getDHCPConfig / getIPAMConfig before starting stateless watcher.
	resp, err := d.client.Get(ctx, KeyIPAM)
	if err != nil {
		return err
	}
	if resp.Count == 0 {
		// not initialized
		return nil
	}

	ipam := new(sabakan.IPAMConfig)
	err = json.Unmarshal(resp.Kvs[0].Value, ipam)
	if err != nil {
		return err
	}

	// copy gateway-offset from dhcp.json to ipam.json
	resp, err = d.client.Get(ctx, KeyDHCP)
	if err != nil {
		return err
	}
	if resp.Count > 0 {
		dc := new(sabakan.DHCPConfig)
		err := json.Unmarshal(resp.Kvs[0].Value, dc)
		if err != nil {
			return err
		}
		ipam.NodeGatewayOffset = dc.GatewayOffset
		data, err := json.Marshal(ipam)
		if err != nil {
			return err
		}
		tresp, err := d.client.Txn(ctx).If(mu.IsOwner()).
			Then(clientv3.OpPut(KeyIPAM, string(data))).
			Commit()
		if err != nil {
			return err
		}
		if !tresp.Succeeded {
			return errLostOwner
		}
	}

	// update Machine.Info
	err = d.convertTo2Machines(ctx, mu, ipam)
	if err != nil {
		return err
	}

	// update schema version
	tresp, err := d.client.Txn(ctx).If(mu.IsOwner()).
		Then(clientv3.OpPut(KeyVersion, sabakan.SchemaVersion)).
		Commit()
	if err != nil {
		return err
	}
	if !tresp.Succeeded {
		return errLostOwner
	}

	log.Info("updated schema version", map[string]interface{}{
		"to": sabakan.SchemaVersion,
	})
	return nil
}
