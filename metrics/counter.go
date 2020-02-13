package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// APICounter represents API counter.
type APICounter struct {
	counter *prometheus.CounterVec
}

// NewCounter returns a new APICounter.
func NewCounter() *APICounter {
	return &APICounter{
		counter: APIRequestTotal,
	}
}

// Inc increments the counter.
func (c *APICounter) Inc(statusCode int, path, verb string) {
	c.counter.WithLabelValues(fmt.Sprint(statusCode), path, verb).Inc()
}
