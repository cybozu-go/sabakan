package mock

import (
	"context"
	"strconv"

	"github.com/cybozu-go/sabakan"
)

func (d *driver) PutTemplate(ctx context.Context, role string, template string) (string, error) {
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

func (d *driver) GetTemplateIDs(ctx context.Context, role string) ([]string, error) {
	templateMap, ok := d.ignitions[role]
	if !ok {
		return nil, sabakan.ErrNotFound
	}
	res := make([]string, 0)
	for k := range templateMap {
		res = append(res, k)
	}
	return res, nil
}

func (d *driver) GetTemplate(ctx context.Context, role string, id string) (string, error) {
	res, ok := d.ignitions[role][id]
	if !ok {
		return "", sabakan.ErrNotFound
	}
	return res, nil
}

func (d *driver) DeleteTemplate(ctx context.Context, role string, id string) error {
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
