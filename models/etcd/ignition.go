package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
)

// PutTemplate implements sabakan.IgnitionModel
func (d *driver) PutTemplate(ctx context.Context, role, id string, template string, metadata map[string]string) error {
	if metadata == nil {
		return errors.New("metadata should not be nil")
	}
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
		return sabakan.ErrConflicted
	}

	d.addLog(ctx, time.Now(), tresp.Header.Revision, sabakan.AuditIgnition, role, "put",
		fmt.Sprintf("id=%s\n%s", id, template))

	tmplPrefix := path.Join(KeyIgnitionsTemplate, role) + "/"
	metaPrefix := path.Join(KeyIgnitionsMetadata, role) + "/"
	resp, err := d.client.Get(ctx, tmplPrefix,
		clientv3.WithPrefix(),
		clientv3.WithKeysOnly())
	if err != nil {
		return err
	}
	if resp.Count <= sabakan.MaxIgnitions {
		return nil
	}

	versions := make([]*version.Version, resp.Count)
	for i, kv := range resp.Kvs {
		id := string(kv.Key[len(tmplPrefix):])
		ver, err := version.NewVersion(id)
		if err != nil {
			return err
		}
		versions[i] = ver
	}

	sort.Sort(version.Collection(versions))

	for _, ver := range versions[:len(versions)-sabakan.MaxIgnitions] {
		_, err = d.client.Txn(ctx).
			Then(
				clientv3.OpDelete(tmplPrefix+ver.Original()),
				clientv3.OpDelete(metaPrefix+ver.Original()),
			).
			Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTemplateIndex implements sabakan.IgnitionModel
func (d *driver) GetTemplateIndex(ctx context.Context, role string) ([]*sabakan.IgnitionInfo, error) {
	target := path.Join(KeyIgnitionsMetadata, role) + "/"
	resp, err := d.client.Get(ctx, target,
		clientv3.WithPrefix(),
	)
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 {
		return nil, sabakan.ErrNotFound
	}

	index := make(map[string]*sabakan.IgnitionInfo)
	versions := make([]*version.Version, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		info := new(sabakan.IgnitionInfo)
		err = json.Unmarshal(kv.Value, &info.Metadata)
		if err != nil {
			return nil, err
		}
		info.ID = string(kv.Key[len(target):])
		index[info.ID] = info

		ver, err := version.NewVersion(info.ID)
		if err != nil {
			return nil, err
		}
		versions[i] = ver
	}

	sort.Sort(version.Collection(versions))

	result := make([]*sabakan.IgnitionInfo, len(versions))
	for i, ver := range versions {
		result[i] = index[ver.Original()]
	}

	return result, nil
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
