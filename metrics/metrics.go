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

