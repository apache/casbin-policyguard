package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// PolicyViolations tracks policy violations by policy name, mode (dry-run/enforce), and result (allowed/denied)
	PolicyViolations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "policywall_policy_violations_total",
			Help: "Total number of policy violations detected",
		},
		[]string{"policy", "mode", "result", "namespace", "resource"},
	)

	// AdmissionRequests tracks all admission requests processed
	AdmissionRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "policywall_admission_requests_total",
			Help: "Total number of admission requests processed",
		},
		[]string{"operation", "namespace", "resource"},
	)

	// PolicyEvaluationDuration tracks time spent evaluating policies
	PolicyEvaluationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "policywall_policy_evaluation_duration_seconds",
			Help:    "Time spent evaluating policies in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"policy"},
	)

	// ActivePolicies tracks the number of active policies
	ActivePolicies = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "policywall_active_policies",
			Help: "Number of active policies loaded",
		},
		[]string{"mode"},
	)
)

func init() {
	// Register custom metrics with controller-runtime's global registry
	metrics.Registry.MustRegister(
		PolicyViolations,
		AdmissionRequests,
		PolicyEvaluationDuration,
		ActivePolicies,
	)
}
