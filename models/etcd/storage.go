package etcd

import (
	"context"
	"errors"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan/v2"
)

// GetEncryptionKey implements sabakan.StorageModel
func (d *driver) GetEncryptionKey(ctx context.Context, serial string, diskByPath string) ([]byte, error) {
	target := path.Join(KeyCrypts, serial, diskByPath)
	resp, err := d.client.Get(ctx, target)
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 {
		return nil, nil
	}

	return resp.Kvs[0].Value, nil
}

// PutEncryptionKey implements sabakan.StorageModel
func (d *driver) PutEncryptionKey(ctx context.Context, serial string, diskByPath string, key []byte) error {
	target := path.Join(KeyCrypts, serial, diskByPath)
	mkey := KeyMachines + serial

RETRY:
	m, rev, err := d.machineGetWithRev(ctx, serial)
	if err != nil {
		return err
	}
	if m.Status.State == sabakan.StateRetiring || m.Status.State == sabakan.StateRetired {
		return errors.New("machine was retiring or retired")
	}

	tresp, err := d.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(mkey), "=", rev)).
		Then(
			clientv3.OpTxn(
				[]clientv3.Cmp{clientv3util.KeyMissing(target)},
				[]clientv3.Op{clientv3.OpPut(target, string(key))},
				nil,
			),
		).
		Commit()
	if err != nil {
		return err
	}

	if !tresp.Succeeded {
		goto RETRY
	}

	if !tresp.Responses[0].GetResponseTxn().Succeeded {
		return sabakan.ErrConflicted
	}

	d.addLog(ctx, time.Now(), tresp.Header.Revision, sabakan.AuditCrypts, serial, "put",
		diskByPath)

	return nil
}

// DeleteEncryptionKeys implements sabakan.StorageModel
func (d *driver) DeleteEncryptionKeys(ctx context.Context, serial string) ([]string, error) {
	mkey := KeyMachines + serial
	ckey := path.Join(KeyCrypts, serial) + "/"

RETRY:
	m, rev, err := d.machineGetWithRev(ctx, serial)
	if err != nil {
		return nil, err
	}
	if m.Status.State != sabakan.StateRetiring {
		return nil, errors.New("machine is not retiring")
	}

	resp, err := d.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(mkey), "=", rev)).
		Then(clientv3.OpDelete(ckey, clientv3.WithPrefix(), clientv3.WithPrevKV())).
		Commit()
	if err != nil {
		return nil, err
	}

	if !resp.Succeeded {
		goto RETRY
	}

	d.addLog(ctx, time.Now(), resp.Header.Revision, sabakan.AuditCrypts, serial, "delete",
		"")

	dresp := resp.Responses[0].GetResponseDeleteRange()

	ret := make([]string, len(dresp.PrevKvs))
	for i, ev := range dresp.PrevKvs {
		ret[i] = string(ev.Key[len(ckey):])
	}
	return ret, nil
}
