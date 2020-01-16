package metrics

import "github.com/prometheus/client_golang/prometheus"

// MachineStatus returns the machine state metrics
var MachineStatus = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "machine_status",
		Help:      "The machine status set by HTTP API.",
	},
	[]string{"status", "address", "serial", "rack", "role", "machine_type"},
)

// APIRequestTotal returns the total count of API calls
var APIRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "api_request_count",
		Help:      "The total count of API calls.",
	},
	[]string{"code", "path", "verb"},
)

// AssetsBytesTotal returns the total bytes of assets
var AssetsBytesTotal = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "assets_bytes_total",
		Help:      "The total bytes of assets.",
	},
)

// AssetsItemsTotal returns the total Items of assets
var AssetsItemsTotal = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "assets_items_total",
		Help:      "The total items of assets.",
	},
)

// ImagesBytesTotal returns the total bytes of Images
var ImagesBytesTotal = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "images_bytes_total",
		Help:      "The total bytes of Images.",
	},
)

// ImagesItemsTotal returns the total Items of Images
var ImagesItemsTotal = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "images_items_total",
		Help:      "The total items of Images.",
	},
)
