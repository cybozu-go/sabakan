package mock

import (
	"context"
	"sort"
	"strconv"
	"sync"

	"github.com/cybozu-go/sabakan"
)

type ignitionDriver struct {
	mu        sync.Mutex
	ignitions map[string]map[string]string
}

func newIgnitionDriver() *ignitionDriver {
	return &ignitionDriver{
		ignitions: make(map[string]map[string]string),
	}
}

func (d *ignitionDriver) PutTemplate(ctx context.Context, role string, template string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	templateMap := d.ignitions[role]
	if templateMap == nil {
		templateMap = make(map[string]string)
		d.ignitions[role] = templateMap
	}
	id := strconv.Itoa(len(templateMap))
	templateMap[id] = template
	return id, nil
}

func (d *ignitionDriver) GetTemplateIDs(ctx context.Context, role string) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	templateMap, ok := d.ignitions[role]
	if !ok {
		return nil, sabakan.ErrNotFound
	}
	res := make([]string, 0)
	for k := range templateMap {
		res = append(res, k)
	}
	sort.Strings(res)
	return res, nil
}

func (d *ignitionDriver) GetTemplate(ctx context.Context, role string, id string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	res, ok := d.ignitions[role][id]
	if !ok {
		return "", sabakan.ErrNotFound
	}
	return res, nil
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
