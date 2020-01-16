package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

const (
	namespace = "sabakan"
)

type logger struct{}

func (l logger) Println(v ...interface{}) {
	log.Error(fmt.Sprint(v...), nil)
}

// GetHandler return http.Handler for prometheus metrics
func GetHandler() http.Handler {
	registry := prometheus.NewRegistry()
	registerMetrics(registry)

	handler := promhttp.HandlerFor(registry,
		promhttp.HandlerOpts{
			ErrorLog:      logger{},
			ErrorHandling: promhttp.ContinueOnError,
		})

	return handler
}

func registerMetrics(registry *prometheus.Registry) {
	registry.MustRegister(MachineStatus)
	registry.MustRegister(AssetsBytesTotal)
	registry.MustRegister(AssetsItemsTotal)
	registry.MustRegister(ImagesBytesTotal)
	registry.MustRegister(ImagesItemsTotal)
}

// Updater updates Prometheus metrics periodically
type Updater struct {
	interval time.Duration
	model    *sabakan.Model
}

// NewUpdater is the constructor for Updater
func NewUpdater(interval time.Duration, model *sabakan.Model) *Updater {
	return &Updater{interval, model}
}

// UpdateLoop is the func to update all metrics continuously
func (u *Updater) UpdateLoop(ctx context.Context) error {
	ticker := time.NewTicker(u.interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := u.UpdateAllMetrics(ctx)
			if err != nil {
				log.Warn("failed to update metrics", map[string]interface{}{
					log.FnError: err.Error(),
				})
			}
		}
	}
}

// UpdateAllMetrics is the func to update all metrics once
func (u *Updater) UpdateAllMetrics(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	tasks := map[string]func(ctx context.Context) error{
		"updateMachineStatus": u.updateMachineStatus,
		"updateAssetMetrics":  u.updateAssetMetrics,
		// "updateImageMetrics":  u.updateImageMetrics,
	}
	for key, task := range tasks {
		key, task := key, task
		g.Go(func() error {
			err := task(ctx)
			if err != nil {
				log.Warn("unable to update metrics", map[string]interface{}{
					"funcname":  key,
					log.FnError: err,
				})
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func (u *Updater) updateMachineStatus(ctx context.Context) error {
	machines, err := u.model.Machine.Query(ctx, nil)
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

func (u *Updater) updateAssetMetrics(ctx context.Context) error {
	assets, err := u.model.Asset.GetInfoAll(ctx)
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

	AssetsBytesTotal.WithLabelValues("byte").Set(float64(totalSize))
	AssetsItemsTotal.WithLabelValues("file").Set(float64(len(assets)))

	return nil
}
