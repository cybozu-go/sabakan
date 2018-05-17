package etcd

import (
	"context"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
)

// GetEncryptionKey implements sabakan.StorageModel
func (d *driver) GetEncryptionKey(ctx context.Context, serial string, diskByPath string) ([]byte, error) {
	target := path.Join(d.prefix, KeyCrypts, serial, diskByPath)
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
	target := path.Join(d.prefix, KeyCrypts, serial, diskByPath)

	tresp, err := d.client.Txn(ctx).
		// Prohibit overwriting
		If(clientv3util.KeyMissing(target)).
		Then(clientv3.OpPut(target, string(key))).
		Else().
		Commit()
	if err != nil {
		return err
	}

	if !tresp.Succeeded {
		return sabakan.ErrConflicted
	}

	return nil
}

// DeleteEncryptionKeys implements sabakan.StorageModel
func (d *driver) DeleteEncryptionKeys(ctx context.Context, serial string) ([]string, error) {
	target := path.Join(d.prefix, KeyCrypts, serial) + "/"

	dresp, err := d.client.Delete(ctx, target, clientv3.WithPrefix(), clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	resp := make([]string, len(dresp.PrevKvs))
	for i, ev := range dresp.PrevKvs {
		resp[i] = string(ev.Key[len(target):])
	}
	return resp, nil
}
