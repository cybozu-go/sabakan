package sabakan

import (
	"context"
	"errors"
	"net"
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

// IPAMModel is an interface for IPAMConfig.
type IPAMModel interface {
	PutConfig(ctx context.Context, config *IPAMConfig) error
	GetConfig(ctx context.Context) (*IPAMConfig, error)
}

// Runner is an interface to run the underlying threads.
type Runner interface {
	Run(ctx context.Context) error
}

// Model is a struct that consists of sub-models.
type Model struct {
	Runner
	Storage StorageModel
	Machine MachineModel
	IPAM    IPAMModel
}

// Leaser is an interface to lease IP addresses
type Leaser interface {
	Lease() (net.IP, error)
}
