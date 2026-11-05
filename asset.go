package sabakan

import "time"

const (
	// AssetOptionAllowedRoles restricts GET and HEAD of the asset to machines whose role is in the comma-separated list.
	AssetOptionAllowedRoles = "allowed-roles"

	// AssetPeerTokenHeader is the HTTP header to send the token that lets sabakan servers pull restricted assets from each other.
	AssetPeerTokenHeader = "X-Sabakan-Asset-Peer-Token"
)

// Asset represents an asset.
type Asset struct {
	Name        string            `json:"name"`
	ID          int               `json:"id,string"`
	ContentType string            `json:"content-type"`
	Date        time.Time         `json:"date"`
	Size        int64             `json:"size"`
	Sha256      string            `json:"sha256"`
	URLs        []string          `json:"urls"`
	Exists      bool              `json:"exists"`
	Options     map[string]string `json:"options"`
}

// AssetStatus is the status of an asset.
type AssetStatus struct {
	Status int `json:"status"`
	ID     int `json:"id,string"`
}
