package sabakan

import "time"

// Asset represents an asset.
type Asset struct {
	Name        string    `json:"name"`
	ID          int       `json:"id,string"`
	ContentType string    `json:"content-type"`
	Date        time.Time `json:"date"`
	Sha256      string    `json:"sha256"`
	URLs        []string  `json:"urls"`
	Version     int64     `json:"version"`
	Exists      bool      `json:"exists"`
}

// AssetStatus is the status of an asset.
type AssetStatus struct {
	Status  int   `json:"status"`
	Version int64 `json:"version"`
	ID      int   `json:"id,string"`
}
