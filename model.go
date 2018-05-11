package sabakan

import (
	"context"
	"errors"
)

// ErrConflicted is a special error for models.
// A model should return this when it fails to update a resource due to conflicts.
var ErrConflicted = errors.New("key conflicted")

// ErrNotFound is a special err for models.
// A model should return this when it cannot find a resource by a specified key.
var ErrNotFound = errors.New("not found")

// StorageModel is an interface for disk encryption keys.
type StorageModel interface {
	GetEncryptionKey(ctx context.Context, serial string, diskByPath string) ([]byte, error)
	PutEncryptionKey(ctx context.Context, serial string, diskByPath string, key []byte) error
	DeleteEncryptionKeys(ctx context.Context, serial string) ([]string, error)
}

// MachineModel is an interface for machine database.
type MachineModel interface {
	Register(ctx context.Context, machines []*Machine) error
	Query(ctx context.Context, query *Query) ([]*Machine, error)
	Delete(ctx context.Context, serial string) error
}

// ConfigModel is an interface for IPAMConfig.
type ConfigModel interface {
	PutConfig(ctx context.Context, config *IPAMConfig) error
	GetConfig(ctx context.Context) (*IPAMConfig, error)
}

// Model is a struct that consists of sub-models.
type Model struct {
	Storage StorageModel
	Machine MachineModel
	Config  ConfigModel
}
