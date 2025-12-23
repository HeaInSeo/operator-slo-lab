package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	E2EConvergenceSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "e2e_convergence_time_seconds",
			Help:    "Time taken from test/start-time annotation to reconciliation completion",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)
)

func init() {
	metrics.Registry.MustRegister(E2EConvergenceSeconds)
}
