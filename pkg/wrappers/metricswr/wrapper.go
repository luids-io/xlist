// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

// Package metricswr provides a wrapper for RBLs that implements prometheus
// metrics.
//
// This package is a work in progress and makes no API stability promises.
package metricswr

import (
	"context"

	cliprom "github.com/prometheus/client_golang/prometheus"

	"github.com/luids-io/api/xlist"
)

// Wrapper implements an xlist.Checker wrapper for metrics
type Wrapper struct {
	listID string
	list   xlist.List
}

// Option is used for component configuration
type Option func(*options)

//stats is a global structure
var stats struct {
	pings     *cliprom.CounterVec
	requests  *cliprom.CounterVec
	durations *cliprom.SummaryVec
}

type options struct{}

// New returns a Wrapper, it recevies the listID used for the metrics
func New(list xlist.List) *Wrapper {
	return &Wrapper{
		list:   list,
		listID: list.ID(),
	}
}

// ID implements xlist.List interface
func (w *Wrapper) ID() string {
	return w.listID
}

// Class implements xlist.List interface
func (w *Wrapper) Class() string {
	return BuildClass
}

// Check implements xlist.Checker interface
func (w *Wrapper) Check(ctx context.Context, name string, resource xlist.Resource) (xlist.Response, error) {
	timer := cliprom.NewTimer(cliprom.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		stats.durations.WithLabelValues(w.listID).Observe(us)
	}))
	defer timer.ObserveDuration()

	resp, err := w.list.Check(ctx, name, resource)
	if err != nil {
		stats.requests.WithLabelValues(w.listID, resource.String(), "fail").Inc()
	} else {
		if resp.Result {
			stats.requests.WithLabelValues(w.listID, resource.String(), "hit").Inc()
		} else {
			stats.requests.WithLabelValues(w.listID, resource.String(), "miss").Inc()
		}
	}
	return resp, err
}

// Ping implements xlist.Checker interface
func (w *Wrapper) Ping() error {
	err := w.list.Ping()
	if err != nil {
		stats.pings.WithLabelValues(w.listID, "fail").Inc()
	} else {
		stats.pings.WithLabelValues(w.listID, "success").Inc()
	}
	return err
}

// Resources implements xlist.Checker interface
func (w *Wrapper) Resources() []xlist.Resource {
	return w.list.Resources()
}

// ReadOnly implements xlist.List interface
func (w *Wrapper) ReadOnly() bool {
	return true
}

func init() {
	stats.pings = cliprom.NewCounterVec(
		cliprom.CounterOpts{
			Name: "xlist_pings_total",
			Help: "How many check pings processed, partitioned by status",
		},
		[]string{"list", "status"})

	stats.requests = cliprom.NewCounterVec(
		cliprom.CounterOpts{
			Name: "xlist_requests_total",
			Help: "How many check requests processed, partitioned by status",
		},
		[]string{"list", "resource", "status"})

	stats.durations = cliprom.NewSummaryVec(
		cliprom.SummaryOpts{
			Name:       "xlist_request_durations",
			Help:       "Xlist request latencies in microseconds",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"list"})

	cliprom.MustRegister(stats.pings)
	cliprom.MustRegister(stats.requests)
	cliprom.MustRegister(stats.durations)
}
