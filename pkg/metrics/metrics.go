// Copyright 2026 The Casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector holds Prometheus metrics
type Collector struct {
	admissionRequests *prometheus.CounterVec
	policyEvaluations *prometheus.HistogramVec
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return NewCollectorWithRegistry(prometheus.DefaultRegisterer)
}

// NewCollectorWithRegistry creates a new metrics collector with a custom registry
func NewCollectorWithRegistry(registerer prometheus.Registerer) *Collector {
	c := &Collector{
		admissionRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "policywall_admission_requests_total",
				Help: "Total number of admission requests",
			},
			[]string{"operation", "allowed"},
		),
		policyEvaluations: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "policywall_policy_evaluation_duration_seconds",
				Help:    "Policy evaluation duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"policy"},
		),
	}

	registerer.MustRegister(c.admissionRequests)
	registerer.MustRegister(c.policyEvaluations)

	return c
}

// RecordAdmissionRequest records an admission request
func (c *Collector) RecordAdmissionRequest(operation string, allowed bool) {
	allowedStr := "true"
	if !allowed {
		allowedStr = "false"
	}
	c.admissionRequests.WithLabelValues(operation, allowedStr).Inc()
}

// Handler returns an HTTP handler for metrics
func (c *Collector) Handler() http.Handler {
	return promhttp.Handler()
}
