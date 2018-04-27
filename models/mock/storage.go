package mock

import (
	"context"
	"path"
	"strings"

	"github.com/cybozu-go/sabakan"
)

// GetEncryptionKey implements sabakan.StorageModel
func (d *driver) GetEncryptionKey(ctx context.Context, serial string, diskByPath string) ([]byte, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	target := path.Join(serial, diskByPath)
	key, ok := d.storage[target]
	if !ok {
		return nil, nil
	}

	return key, nil
}

// PutEncryptionKey implements sabakan.StorageModel
func (d *driver) PutEncryptionKey(ctx context.Context, serial string, diskByPath string, key []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	target := path.Join(serial, diskByPath)
	_, ok := d.storage[target]
	if ok {
		return sabakan.ErrConflicted
	}
	d.storage[target] = key

	return nil
}

// DeleteEncryptionKeys implements sabakan.StorageModel
func (d *driver) DeleteEncryptionKeys(ctx context.Context, serial string) ([]string, error) {
	prefix := serial + "/"

	d.mu.Lock()
	defer d.mu.Unlock()

	var resp []string
	for k := range d.storage {
		if strings.HasPrefix(k, prefix) {
			delete(d.storage, k)
			resp = append(resp, k)
		}
	}

	return resp, nil
}
