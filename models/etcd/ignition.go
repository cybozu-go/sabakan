package etcd

import (
	"context"
	"path"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
)

// PutTemplate implements sabakan.IgnitionModel
func (d *driver) PutTemplate(ctx context.Context, role string, template string) (string, error) {
RETRY:
	now := time.Now()
	id := strconv.FormatInt(now.UnixNano(), 10)
	target := path.Join(d.prefix, KeyIgnitions, role, id)

	tresp, err := d.client.Txn(ctx).
		// Prohibit overwriting
		If(clientv3util.KeyMissing(target)).
		Then(clientv3.OpPut(target, template)).
		Else().
		Commit()
	if err != nil {
		return "", err
	}

	if !tresp.Succeeded {
		time.Sleep(1 * time.Nanosecond)
		goto RETRY
	}

	target = path.Join(d.prefix, KeyIgnitions, role) + "/"
	resp, err := d.client.Get(ctx, target,
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return "", err
	}
	if resp.Count <= sabakan.MaxIgnitions {
		return id, nil
	}

	ops := make([]clientv3.Op, resp.Count-sabakan.MaxIgnitions)
	for i := 0; i < len(ops); i++ {
		idx := i + int(resp.Count-sabakan.MaxIgnitions)
		ops[i] = clientv3.OpDelete(string(resp.Kvs[idx].Key))
	}
	_, err = d.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return "", err
	}

	return id, nil
}

// GetTemplateIDs implements sabakan.IgnitionModel
func (d *driver) GetTemplateIDs(ctx context.Context, role string) ([]string, error) {
	target := path.Join(d.prefix, KeyIgnitions, role)
	resp, err := d.client.Get(ctx, target, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 {
		return nil, sabakan.ErrNotFound
	}

	ids := make([]string, len(resp.Kvs))
	for i, v := range resp.Kvs {
		id := v.Key[len(target):]
		ids[i] = string(id)
	}

	return ids, nil
}

// GetTemplate implements sabakan.IgnitionModel
func (d *driver) GetTemplate(ctx context.Context, role string, id string) (string, error) {
	target := path.Join(d.prefix, KeyIgnitions, role, id)
	resp, err := d.client.Get(ctx, target)
	if err != nil {
		return "", err
	}

	if resp.Count == 0 {
		return "", sabakan.ErrNotFound
	}

	return string(resp.Kvs[0].Value), nil
}

// DeleteTemplate implements sabakan.IgnitionModel
func (d *driver) DeleteTemplate(ctx context.Context, role string, id string) error {
	target := path.Join(d.prefix, KeyIgnitions, role, id)
	resp, err := d.client.Delete(ctx, target)
	if err != nil {
		return err
	}

	if resp.Deleted == 0 {
		return sabakan.ErrNotFound
	}

	return nil
}
