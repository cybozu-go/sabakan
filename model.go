package sabakan

import (
	"context"
	"errors"
)

// ErrConflicted is a special error for models.
// A model should return this when it fails to update a resouce due to conflicts.
var ErrConflicted = errors.New("key conflicted")

// StorageModel is an interface for disk encryption keys.
type StorageModel interface {
	GetEncryptionKey(ctx context.Context, serial string, diskByPath string) ([]byte, error)
	PutEncryptionKey(ctx context.Context, serial string, diskByPath string, key []byte) error
	DeleteEncryptionKeys(ctx context.Context, serial string) ([]string, error)
}

// Model is a struct that consists of sub-models.
type Model struct {
	Storage StorageModel
}
