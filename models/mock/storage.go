package mock

import (
	"context"
	"path"
	"strings"
	"sync"

	"github.com/cybozu-go/sabakan"
)

type storageDriver struct {
	mu      sync.Mutex
	storage map[string][]byte
}

func newStorageDriver() *storageDriver {
	return &storageDriver{
		storage: make(map[string][]byte),
	}
}

// GetEncryptionKey implements sabakan.StorageModel
func (d *storageDriver) GetEncryptionKey(ctx context.Context, serial string, diskByPath string) ([]byte, error) {
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
func (d *storageDriver) PutEncryptionKey(ctx context.Context, serial string, diskByPath string, key []byte) error {
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
func (d *storageDriver) DeleteEncryptionKeys(ctx context.Context, serial string) ([]string, error) {
	prefix := serial + "/"

	d.mu.Lock()
	defer d.mu.Unlock()

	var resp []string
	for k := range d.storage {
		if strings.HasPrefix(k, prefix) {
			delete(d.storage, k)
			resp = append(resp, k[len(serial)+1:])
		}
	}

	return resp, nil
}
