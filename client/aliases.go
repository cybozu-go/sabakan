package client

import "github.com/cybozu-go/sabakan"

// ImageIndex is a list of *Image.
type ImageIndex = sabakan.ImageIndex

// IgnitionVersion represents ignition specification version in `major.minor`
type IgnitionVersion = sabakan.IgnitionVersion

// Supported ignition versions
const (
	Ignition2_3 = sabakan.Ignition2_3
)

// IgnitionTemplate represents an ignition template
type IgnitionTemplate = sabakan.IgnitionTemplate
