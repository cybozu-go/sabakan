package etcd

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
