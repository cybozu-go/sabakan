package etcd

// Internal schema keys.
const (
	KeyCrypts      = "/crypts"
	KeyDHCP        = "/dhcp"
	KeyIPAM        = "/ipam"
	KeyLeaseUsages = "/lease-usages"
	KeyMachines    = "/machines"
	KeyNodeIndices = "/node-indices"
	KeyImages      = "/images"
)

// MaxDeleted is the maximum number of deleted image IDs stored in etcd.
const MaxDeleted = 10
