package sabakan

// DefaultTestConfig is configuration for test; can be used in other packages
var DefaultTestConfig = IPAMConfig{
	MaxRacks:        80,
	MaxNodesInRack:  28,
	NodeIPv4Offset:  "10.69.0.0/26",
	NodeRackShift:   6,
	NodeIndexOffset: 3,
	BMCIPv4Offset:   "10.72.17.0/27",
	BMCRackShift:    5,
	NodeIPPerNode:   3,
	BMCIPPerNode:    1,
}
