package metrics

import "github.com/prometheus/client_golang/prometheus"

// MachineStatus returns the machine state metrics
var MachineStatus = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "machine_status",
		Help:      "The machine status set by HTTP API.",
	},
	[]string{"status", "address", "serial", "rack", "index"},
)

// APIRequestTotal returns the total count of API calls
var APIRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "api_request_count",
		Help:      "The total count of API calls.",
	},
	[]string{},
)

// AssetsBytesTotal returns the total bytes of assets
var AssetsBytesTotal = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "assets_bytes_total",
		Help:      "The total bytes of assets.",
	},
	[]string{"hoge"},
)

// AssetsItemsTotal returns the total Items of assets
var AssetsItemsTotal = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "assets_items_total",
		Help:      "The total items of assets.",
	},
	[]string{"hoge"},
)

// ImagesBytesTotal returns the total bytes of Images
var ImagesBytesTotal = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "images_bytes_total",
		Help:      "The total bytes of Images.",
	},
	[]string{},
)

// ImagesItemsTotal returns the total Items of Images
var ImagesItemsTotal = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "images_items_total",
		Help:      "The total items of Images.",
	},
	[]string{},
)
