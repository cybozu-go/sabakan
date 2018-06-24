package etcd

import "time"

// Internal schema keys.
const (
	KeyCrypts      = "crypts/"
	KeyDHCP        = "dhcp"
	KeyIPAM        = "ipam"
	KeyLeaseUsages = "lease-usages/"
	KeyMachines    = "machines/"
	KeyNodeIndices = "node-indices/"
	KeyImages      = "images/"
	KeyAssets      = "assets/"
	KeyAssetsID    = "assets"
	KeyIgnitions   = "ignitions/"
	KeyAudit       = "audit/"
	KeyAuditLastGC = "audit"
)

// MaxDeleted is the maximum number of deleted image IDs stored in etcd.
const MaxDeleted = 10

// LastRevFile is the filename that keeps the last revision that
// the stateful watcher processed successfully.
const LastRevFile = "lastrev"

// Miscellaneous
const (
	assetPageSize    = 100
	maxJitterSeconds = 30
	maxAssetURLs     = 10
	maxImageURLs     = 10
)

// Log parameters
const (
	logRetentionDays      = 60
	logCompactionTick     = 1 * time.Hour
	logCompactionInterval = 23 * time.Hour
	logPageSize           = 100
)
