package etcd

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/cybozu-go/sabakan/v2"
	version "github.com/hashicorp/go-version"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/clientv3util"
)

func keyIgnitionRolePrefix(role string) string {
	return KeyIgnitions + role + "/"
}

func (d *driver) PutTemplate(ctx context.Context, role, id string, tmpl *sabakan.IgnitionTemplate) error {
	pfx := keyIgnitionRolePrefix(role)
	target := pfx + id
	data, err := json.Marshal(tmpl)
	if err != nil {
		return err
	}

	tresp, err := d.client.Txn(ctx).
		// Prohibit overwriting
		If(clientv3util.KeyMissing(target)).
		Then(clientv3.OpPut(target, string(data))).
		Commit()
	if err != nil {
		return err
	}

	if !tresp.Succeeded {
		return sabakan.ErrConflicted
	}

	d.addLog(ctx, time.Now(), tresp.Header.Revision, sabakan.AuditIgnition, role, "put", id)

	resp, err := d.client.Get(ctx, pfx, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return err
	}
	if resp.Count <= MaxIgnitions {
		return nil
	}

	versions := make([]*version.Version, resp.Count)
	for i, kv := range resp.Kvs {
		ver, err := version.NewVersion(string(kv.Key[len(pfx):]))
		if err != nil {
			return err
		}
		versions[i] = ver
	}

	sort.Sort(version.Collection(versions))

	for _, ver := range versions[:len(versions)-MaxIgnitions] {
		_, err = d.client.Delete(ctx, pfx+ver.Original())
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *driver) GetTemplateIDs(ctx context.Context, role string) ([]string, error) {
	pfx := keyIgnitionRolePrefix(role)
	resp, err := d.client.Get(ctx, pfx, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, nil
	}

	versions := make([]*version.Version, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		id := string(kv.Key[len(pfx):])
		ver, err := version.NewVersion(id)
		if err != nil {
			return nil, err
		}
		versions[i] = ver
	}

	sort.Sort(version.Collection(versions))

	result := make([]string, len(versions))
	for i, ver := range versions {
		result[i] = ver.Original()
	}

	return result, nil
}

func (d *driver) GetTemplate(ctx context.Context, role string, id string) (*sabakan.IgnitionTemplate, error) {
	target := keyIgnitionRolePrefix(role) + id
	resp, err := d.client.Get(ctx, target)
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 {
		return nil, sabakan.ErrNotFound
	}

	tmpl := new(sabakan.IgnitionTemplate)
	err = json.Unmarshal(resp.Kvs[0].Value, tmpl)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// DeleteTemplate implements sabakan.IgnitionModel
func (d *driver) DeleteTemplate(ctx context.Context, role string, id string) error {
	target := keyIgnitionRolePrefix(role) + id
	resp, err := d.client.Delete(ctx, target)
	if err != nil {
		return err
	}
	if resp.Deleted == 0 {
		return sabakan.ErrNotFound
	}

	d.addLog(ctx, time.Now(), resp.Header.Revision, sabakan.AuditIgnition, role, "delete", id)

	return nil
}
