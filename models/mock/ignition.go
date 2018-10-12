package mock

import (
	"context"
	"sort"
	"strconv"
	"sync"

	"github.com/cybozu-go/sabakan"
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

func (d *ignitionDriver) PutTemplate(ctx context.Context, role string, template string, metadata map[string]string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	templateMap := d.ignitions[role]
	if templateMap == nil {
		templateMap = make(map[string]ignitionData)
		d.ignitions[role] = templateMap
	}
	id := strconv.Itoa(len(templateMap))
	templateMap[id] = ignitionData{
		template: template,
		metadata: metadata,
	}
	return id, nil
}

func (d *ignitionDriver) GetTemplateMetadataList(ctx context.Context, role string) ([]map[string]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	templateMap, ok := d.ignitions[role]
	if !ok {
		return nil, sabakan.ErrNotFound
	}
	res := make([]map[string]string, 0)
	for k, v := range templateMap {
		meta := map[string]string{
			"id": k,
		}
		for k2, v2 := range v.metadata {
			meta[k2] = v2
		}
		res = append(res, meta)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i]["id"] < res[j]["id"]
	})
	return res, nil
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
