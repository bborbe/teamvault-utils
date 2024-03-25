// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// NewMetrics create prometheus metrics for the given Func.
func NewMetrics(
	registerer prometheus.Registerer,
	namespace string,
	subsystem string,
	fn Func,
) func(ctx context.Context) error {
	started := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "started",
		Help:      "started",
	})
	completed := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "completed",
		Help:      "completed",
	})
	failed := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "failed",
		Help:      "failed",
	})
	lastSuccess := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "last_success",
		Help:      "Timestamp of last successful run",
	})
	registerer.MustRegister(
		started,
		completed,
		failed,
		lastSuccess,
	)
	return func(ctx context.Context) error {
		started.Inc()
		if err := fn(ctx); err != nil {
			failed.Inc()
			return err
		}
		completed.Inc()
		lastSuccess.SetToCurrentTime()
		return nil
	}
}
