package mock

import (
	"context"
	"sort"
	"sync"

	"github.com/cybozu-go/sabakan"
	version "github.com/hashicorp/go-version"
)

type ignitionData struct {
	template string
	metadata map[string]string
}

type ignitionDriver struct {
	mu        sync.Mutex
	ignitions map[string]map[string]ignitionData
}

func newIgnitionDriver() *ignitionDriver {
	return &ignitionDriver{
		ignitions: make(map[string]map[string]ignitionData),
	}
}

func (d *ignitionDriver) PutTemplate(ctx context.Context, role, id string, template string, metadata map[string]string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	templateMap := d.ignitions[role]
	if templateMap == nil {
		templateMap = make(map[string]ignitionData)
		d.ignitions[role] = templateMap
	}
	templateMap[id] = ignitionData{
		template: template,
		metadata: metadata,
	}
	return nil
}

func (d *ignitionDriver) GetTemplateIndex(ctx context.Context, role string) ([]*sabakan.IgnitionInfo, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	templateMap, ok := d.ignitions[role]
	if !ok {
		return nil, sabakan.ErrNotFound
	}

	versions := make([]*version.Version, len(templateMap))
	i := 0
	for k := range templateMap {
		ver, err := version.NewVersion(k)
		if err != nil {
			return nil, err
		}
		versions[i] = ver
		i++
	}

	sort.Sort(version.Collection(versions))

	result := make([]*sabakan.IgnitionInfo, len(versions))
	for i, ver := range versions {
		id := ver.Original()
		info := &sabakan.IgnitionInfo{
			ID:       id,
			Metadata: templateMap[id].metadata,
		}
		result[i] = info
	}

	return result, nil
}

func (d *ignitionDriver) GetTemplate(ctx context.Context, role string, id string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	res, ok := d.ignitions[role][id]
	if !ok {
		return "", sabakan.ErrNotFound
	}
	return res.template, nil
}

func (d *ignitionDriver) GetTemplateMetadata(ctx context.Context, role string, id string) (map[string]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	res, ok := d.ignitions[role][id]
	if !ok {
		return nil, sabakan.ErrNotFound
	}
	return res.metadata, nil
}

func (d *ignitionDriver) DeleteTemplate(ctx context.Context, role string, id string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	ids, ok := d.ignitions[role]
	if !ok {
		return sabakan.ErrNotFound
	}
	if _, ok := ids[id]; !ok {
		return sabakan.ErrNotFound
	}
	delete(ids, id)
	return nil
}
