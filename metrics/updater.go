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

func (u *Updater) UpdateLoop(ctx context.Context) error {
	err := u.updateAllMetrics(ctx)
	if err != nil {
		log.Warn("failed to update metrics", map[string]interface{}{
			log.FnError: err.Error(),
		})
	}
	ticker := time.NewTicker(u.interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := u.updateAllMetrics(ctx)
			if err != nil {
				log.Warn("failed to update metrics", map[string]interface{}{
					log.FnError: err.Error(),
				})
			}
		}
	}
}

func (u *Updater) updateAllMetrics(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	tasks := map[string]func(ctx context.Context) error{
		"updateMachineStatus": u.updateMachineStatus,
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
		for _, st := range sabakan.StateList {
			labelValues := []string{st.String(), m.Spec.IPv4[0], m.Spec.Serial, fmt.Sprint(m.Spec.Rack), fmt.Sprint(m.Spec.IndexInRack)}
			if st == m.Status.State {
				MachineStatus.WithLabelValues(labelValues...).Set(1)
			} else {
				MachineStatus.WithLabelValues(labelValues...).Set(0)
			}
		}
	}
	return nil
}
