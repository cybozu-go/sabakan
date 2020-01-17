package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// APICounter represents API counter.
type APICounter struct {
	Counter *prometheus.CounterVec
}

// NewCounter returns a new APICounter.
func NewCounter() *APICounter {
	return &APICounter{
		Counter: APIRequestTotal,
	}
}

// Inc increments APIRequestTotal counter
func (c *APICounter) Inc(statusCode int, path, verb string) {
	c.Counter.WithLabelValues(fmt.Sprint(statusCode), path, verb).Inc()
}
