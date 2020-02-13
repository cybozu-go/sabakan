package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace     = "sabakan"
	scrapeTimeout = time.Second * 10
)

type logger struct{}

func (l logger) Println(v ...interface{}) {
	log.Error(fmt.Sprint(v...), nil)
}

// Metric represents collectors and updater of metric.
type Metric struct {
	collectors []prometheus.Collector
	updater    func(context.Context, *sabakan.Model) error
}

// Collector is a metrics collector for Sabakan.
type Collector struct {
	metrics map[string]Metric
	model   *sabakan.Model
}

// NewCollector returns a new Collector.
func NewCollector(model *sabakan.Model) *Collector {
	return &Collector{
		metrics: map[string]Metric{
			"machine_status": {
				collectors: []prometheus.Collector{MachineStatus},
				updater:    updateMachineStatus,
			},
			"api_request_count": {
				collectors: []prometheus.Collector{APIRequestTotal},
				updater:    updateNop,
			},
			"assets_total": {
				collectors: []prometheus.Collector{AssetsBytesTotal, AssetsItemsTotal},
				updater:    updateAssetMetrics,
			},
			"images_total": {
				collectors: []prometheus.Collector{ImagesBytesTotal, ImagesItemsTotal},
				updater:    updateImageMetrics,
			},
		},
		model: model,
	}
}

// GetHandler return http.Handler for prometheus metrics
func GetHandler(collector *Collector) http.Handler {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	handler := promhttp.HandlerFor(registry,
		promhttp.HandlerOpts{
			ErrorLog:      logger{},
			ErrorHandling: promhttp.ContinueOnError,
		})

	return handler
}

// Describe implements prometheus.Collector.Describe().
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		for _, col := range metric.collectors {
			col.Describe(ch)
		}
	}
}

// Collect implements prometheus.Collector.Collect().
func (c Collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), scrapeTimeout)
	defer cancel()

	var wg sync.WaitGroup
	for key, metric := range c.metrics {
		wg.Add(1)
		go func(key string, metric Metric) {
			defer wg.Done()
			err := metric.updater(ctx, c.model)
			if err != nil {
				log.Warn("unable to update metrics", map[string]interface{}{
					"name":      key,
					log.FnError: err,
				})
				return
			}

			for _, col := range metric.collectors {
				col.Collect(ch)
			}
		}(key, metric)
	}
	wg.Wait()
}

func updateMachineStatus(ctx context.Context, model *sabakan.Model) error {
	machines, err := model.Machine.Query(ctx, nil)
	if err != nil {
		return err
	}

	for _, m := range machines {
		if len(m.Spec.IPv4) == 0 {
			return fmt.Errorf("unable to expose metrics, because machine have no IPv4 address; serial: %s", m.Spec.Serial)
		}
		for _, st := range sabakan.StateList {
			labelValues := []string{st.String(), m.Spec.IPv4[0], m.Spec.Serial, fmt.Sprint(m.Spec.Rack), m.Spec.Role, m.Spec.Labels["machine-type"]}
			if st == m.Status.State {
				MachineStatus.WithLabelValues(labelValues...).Set(1)
			} else {
				MachineStatus.WithLabelValues(labelValues...).Set(0)
			}
		}
	}

	return nil
}

func updateAssetMetrics(ctx context.Context, model *sabakan.Model) error {
	assets, err := model.Asset.GetInfoAll(ctx)
	if err != nil {
		return err
	}
	if len(assets) == 0 {
		return nil
	}

	var totalSize int64
	for _, a := range assets {
		totalSize += a.Size
	}

	AssetsBytesTotal.Set(float64(totalSize))
	AssetsItemsTotal.Set(float64(len(assets)))

	return nil
}

func updateImageMetrics(ctx context.Context, model *sabakan.Model) error {
	images, err := model.Image.GetInfoAll(ctx)
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return nil
	}

	var totalSize int64
	for _, i := range images {
		totalSize += i.Size
	}

	ImagesBytesTotal.Set(float64(totalSize))
	ImagesItemsTotal.Set(float64(len(images)))

	return nil
}

func updateNop(_ context.Context, _ *sabakan.Model) error {
	return nil
}
