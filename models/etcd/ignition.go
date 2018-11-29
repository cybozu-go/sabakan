package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
	"github.com/pkg/errors"
)

// PutTemplate implements sabakan.IgnitionModel
func (d *driver) PutTemplate(ctx context.Context, role, id string, template string, metadata map[string]string) error {
	if metadata == nil {
		return errors.New("metadata should not be nil")
	}
RETRY:
	now := time.Now()
	target := path.Join(KeyIgnitionsTemplate, role, id)
	meta := path.Join(KeyIgnitionsMetadata, role, id)
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	tresp, err := d.client.Txn(ctx).
		// Prohibit overwriting
		If(clientv3util.KeyMissing(target), clientv3util.KeyMissing(meta)).
		Then(clientv3.OpPut(target, template), clientv3.OpPut(meta, string(metaJSON))).
		Else().
		Commit()
	if err != nil {
		return err
	}

	if !tresp.Succeeded {
		time.Sleep(1 * time.Nanosecond)
		goto RETRY
	}

	d.addLog(ctx, now, tresp.Header.Revision, sabakan.AuditIgnition, role, "put",
		fmt.Sprintf("id=%s\n%s", id, template))

	tmplPrefix := path.Join(KeyIgnitionsTemplate, role) + "/"
	resp, err := d.client.Get(ctx, tmplPrefix,
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err
	}
	if resp.Count <= sabakan.MaxIgnitions {
		return nil
	}
	tmplEnd := string(resp.Kvs[resp.Count-sabakan.MaxIgnitions].Key)

	metaPrefix := path.Join(KeyIgnitionsMetadata, role) + "/"
	resp, err = d.client.Get(ctx, metaPrefix,
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err
	}
	if resp.Count <= sabakan.MaxIgnitions {
		return nil
	}
	metaEnd := string(resp.Kvs[resp.Count-sabakan.MaxIgnitions].Key)

	tresp, err = d.client.Txn(ctx).
		Then(
			clientv3.OpDelete(tmplPrefix, clientv3.WithRange(tmplEnd)),
			clientv3.OpDelete(metaPrefix, clientv3.WithRange(metaEnd)),
		).
		Commit()
	if err != nil {
		return err
	}

	return nil
}

// GetTemplateMetadataList implements sabakan.IgnitionModel
func (d *driver) GetTemplateMetadataList(ctx context.Context, role string) ([]map[string]string, error) {
	target := path.Join(KeyIgnitionsMetadata, role) + "/"
	resp, err := d.client.Get(ctx, target,
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
	)
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 {
		return nil, sabakan.ErrNotFound
	}

	metadata := make([]map[string]string, len(resp.Kvs))
	for i, v := range resp.Kvs {
		meta := make(map[string]string)
		err = json.Unmarshal(v.Value, &meta)
		if err != nil {
			return nil, err
		}
		meta["id"] = string(v.Key[len(target):])
		metadata[i] = meta
	}

	return metadata, nil
}

// GetTemplate implements sabakan.IgnitionModel
func (d *driver) GetTemplate(ctx context.Context, role string, id string) (string, error) {
	target := path.Join(KeyIgnitionsTemplate, role, id)
	resp, err := d.client.Get(ctx, target)
	if err != nil {
		return "", err
	}

	if resp.Count == 0 {
		return "", sabakan.ErrNotFound
	}

	return string(resp.Kvs[0].Value), nil
}

// GetTemplateMetadata implements sabakan.IgnitionModel
func (d *driver) GetTemplateMetadata(ctx context.Context, role string, id string) (map[string]string, error) {
	target := path.Join(KeyIgnitionsMetadata, role, id)
	resp, err := d.client.Get(ctx, target)
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 {
		return nil, sabakan.ErrNotFound
	}
	var metadata map[string]string
	err = json.Unmarshal(resp.Kvs[0].Value, &metadata)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

// DeleteTemplate implements sabakan.IgnitionModel
func (d *driver) DeleteTemplate(ctx context.Context, role string, id string) error {
	tmplTarget := path.Join(KeyIgnitionsTemplate, role, id)
	metaTarget := path.Join(KeyIgnitionsMetadata, role, id)
	tresp, err := d.client.Txn(ctx).
		Then(
			clientv3.OpDelete(tmplTarget),
			clientv3.OpDelete(metaTarget),
		).
		Commit()
	if err != nil {
		return err
	}

	if tresp.Responses[0].GetResponseDeleteRange().Deleted == 0 || tresp.Responses[1].GetResponseDeleteRange().Deleted == 0 {
		return sabakan.ErrNotFound
	}

	d.addLog(ctx, time.Now(), tresp.Header.Revision, sabakan.AuditIgnition, role, "delete", id)

	return nil
}
