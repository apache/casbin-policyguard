package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/policywall/api/v1alpha1"
	"github.com/casbin/policywall/pkg/metrics"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func init() {
	_ = admissionv1.AddToScheme(scheme)
}

// ViolationCallback is called when a violation is detected
type ViolationCallback func(policyName string, violation v1alpha1.ViolationResource)

// PolicyEnforcer represents a Casbin enforcer with dry-run configuration
type PolicyEnforcer struct {
	Name     string
	Enforcer *casbin.Enforcer
	DryRun   bool
	Rules    []v1alpha1.MatchRule
}

// WebhookServer handles admission requests and enforces policies
type WebhookServer struct {
	mu                sync.RWMutex
	enforcers         map[string]*PolicyEnforcer
	violationCallback ViolationCallback
}

// NewWebhookServer creates a new webhook server
func NewWebhookServer() *WebhookServer {
	return &WebhookServer{
		enforcers: make(map[string]*PolicyEnforcer),
	}
}

// SetViolationCallback sets the callback for violation reporting
func (s *WebhookServer) SetViolationCallback(cb ViolationCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.violationCallback = cb
}

// UpdatePolicy updates or creates a policy enforcer
func (s *WebhookServer) UpdatePolicy(policy *v1alpha1.AdmissionPolicy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse the Casbin model
	m, err := model.NewModelFromString(policy.Spec.Model)
	if err != nil {
		return fmt.Errorf("failed to parse model: %v", err)
	}

	// Create enforcer
	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return fmt.Errorf("failed to create enforcer: %v", err)
	}

	// Load policies
	policies := parsePolicy(policy.Spec.Policy)
	for _, p := range policies {
		if len(p) > 0 {
			params := make([]interface{}, len(p))
			for i, v := range p {
				params[i] = v
			}
			_, err := enforcer.AddPolicy(params...)
			if err != nil {
				klog.Warningf("Failed to add policy: %v", err)
			}
		}
	}

	// Load role assignments (g rules)
	roleAssignments := parseRoles(policy.Spec.Policy)
	for _, r := range roleAssignments {
		if len(r) > 0 {
			params := make([]interface{}, len(r))
			for i, v := range r {
				params[i] = v
			}
			_, err := enforcer.AddGroupingPolicy(params...)
			if err != nil {
				klog.Warningf("Failed to add grouping policy: %v", err)
			}
		}
	}

	s.enforcers[policy.Name] = &PolicyEnforcer{
		Name:     policy.Name,
		Enforcer: enforcer,
		DryRun:   policy.Spec.DryRun,
		Rules:    policy.Spec.MatchRules,
	}

	// Update metrics
	mode := "enforce"
	if policy.Spec.DryRun {
		mode = "dryrun"
	}
	metrics.ActivePolicies.WithLabelValues(mode).Inc()

	klog.Infof("Updated policy %s (dryRun: %v)", policy.Name, policy.Spec.DryRun)
	return nil
}

// DeletePolicy removes a policy enforcer
func (s *WebhookServer) DeletePolicy(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pe, exists := s.enforcers[name]; exists {
		mode := "enforce"
		if pe.DryRun {
			mode = "dryrun"
		}
		metrics.ActivePolicies.WithLabelValues(mode).Dec()
	}

	delete(s.enforcers, name)
	klog.Infof("Deleted policy %s", name)
}

// ServeHTTP handles admission review requests
func (s *WebhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Invalid content type: %s", contentType)
		http.Error(w, "invalid Content-Type, expect application/json", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *admissionv1.AdmissionResponse
	ar := admissionv1.AdmissionReview{}
	if _, _, err := codecs.UniversalDeserializer().Decode(body, nil, &ar); err != nil {
		klog.Errorf("Can't decode body: %v", err)
		admissionResponse = &admissionv1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		admissionResponse = s.handleAdmission(&ar)
	}

	admissionReview := admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
	}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

// handleAdmission processes the admission request
func (s *WebhookServer) handleAdmission(ar *admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	req := ar.Request

	klog.Infof("AdmissionReview for Kind=%s, Namespace=%s Name=%s UID=%s Operation=%s",
		req.Kind.Kind, req.Namespace, req.Name, req.UID, req.Operation)

	// Record admission request metrics
	metrics.AdmissionRequests.WithLabelValues(
		string(req.Operation),
		req.Namespace,
		req.Kind.Kind,
	).Inc()

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all violations
	var violations []string
	var denyReasons []string

	// Check all applicable policies
	for _, pe := range s.enforcers {
		if !s.matchesRules(req, pe.Rules) {
			continue
		}

		// Track policy evaluation time
		startTime := time.Now()

		// Build enforcement parameters
		params := s.buildEnforcementParams(req)

		// Check policy enforcement
		allowed, err := pe.Enforcer.Enforce(params...)

		// Record evaluation duration
		metrics.PolicyEvaluationDuration.WithLabelValues(pe.Name).Observe(time.Since(startTime).Seconds())

		if err != nil {
			klog.Errorf("Error enforcing policy %s: %v", pe.Name, err)
			continue
		}

		if !allowed {
			// Build detailed violation message with rule context
			user := req.UserInfo.Username
			if user == "" {
				user = "unknown"
			}

			violationMsg := fmt.Sprintf("Policy '%s' denied: user '%s' cannot %s %s '%s/%s' in namespace '%s'. Required: policy rule matching (subject=%s, object=%s/%s, action=%s)",
				pe.Name,
				user,
				req.Operation,
				req.Kind.Kind,
				req.Namespace,
				req.Name,
				req.Namespace,
				user,
				req.Namespace,
				req.Name,
				req.Operation,
			)

			mode := "dryrun"
			result := "allowed"

			if pe.DryRun {
				// In dry-run mode: log and add to warnings
				klog.Warningf("[DRY-RUN] %s", violationMsg)
				violations = append(violations, violationMsg)

				// Create violation resource for status tracking
				violation := v1alpha1.ViolationResource{
					Kind:      req.Kind.Kind,
					Namespace: req.Namespace,
					Name:      req.Name,
					Operation: string(req.Operation),
					Timestamp: metav1.Now(),
					User:      user,
					Message:   violationMsg,
				}

				// Report violation via callback
				if s.violationCallback != nil {
					s.violationCallback(pe.Name, violation)
				}
			} else {
				// Not in dry-run mode: this is a real denial
				mode = "enforce"
				result = "denied"
				denyReasons = append(denyReasons, violationMsg)
			}

			// Record violation metrics
			metrics.PolicyViolations.WithLabelValues(
				pe.Name,
				mode,
				result,
				req.Namespace,
				req.Kind.Kind,
			).Inc()
		}
	}

	// Build response
	if len(denyReasons) > 0 {
		// At least one non-dry-run policy denied the request
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status:  "Failure",
				Message: fmt.Sprintf("Request denied by policies: %v", denyReasons),
				Reason:  metav1.StatusReasonForbidden,
			},
		}
	}

	// Either all policies allowed, or only dry-run violations occurred
	response := &admissionv1.AdmissionResponse{
		Allowed: true,
	}

	// Add violations as warnings
	if len(violations) > 0 {
		response.Warnings = violations
	}

	return response
}

// matchesRules checks if the request matches the policy's rules
func (s *WebhookServer) matchesRules(req *admissionv1.AdmissionRequest, rules []v1alpha1.MatchRule) bool {
	if len(rules) == 0 {
		return true // No rules means match all
	}

	for _, rule := range rules {
		if s.matchesRule(req, rule) {
			return true
		}
	}
	return false
}

// matchesRule checks if a request matches a single rule
func (s *WebhookServer) matchesRule(req *admissionv1.AdmissionRequest, rule v1alpha1.MatchRule) bool {
	// Check API groups
	if len(rule.APIGroups) > 0 && !contains(rule.APIGroups, req.Kind.Group) && !contains(rule.APIGroups, "*") {
		return false
	}

	// Check API versions
	if len(rule.APIVersions) > 0 && !contains(rule.APIVersions, req.Kind.Version) && !contains(rule.APIVersions, "*") {
		return false
	}

	// Check resources
	if len(rule.Resources) > 0 && !contains(rule.Resources, req.Resource.Resource) && !contains(rule.Resources, "*") {
		return false
	}

	// Check operations
	if len(rule.Operations) > 0 && !contains(rule.Operations, string(req.Operation)) && !contains(rule.Operations, "*") {
		return false
	}

	return true
}

// buildEnforcementParams builds Casbin enforcement parameters from admission request
func (s *WebhookServer) buildEnforcementParams(req *admissionv1.AdmissionRequest) []interface{} {
	// Default RBAC-style enforcement: subject, resource, action
	subject := req.UserInfo.Username
	resource := fmt.Sprintf("%s/%s", req.Namespace, req.Name)
	action := string(req.Operation)

	return []interface{}{subject, resource, action}
}

// parsePolicy parses policy string into a list of policy rules
func parsePolicy(policyStr string) [][]string {
	var policies [][]string
	lines := splitLines(policyStr)

	for _, line := range lines {
		line = trimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := splitComma(line)
		if len(parts) > 0 && parts[0] == "p" {
			// Skip the policy type prefix
			parts = parts[1:]
			if len(parts) > 0 {
				policies = append(policies, parts)
			}
		}
	}

	return policies
}

// parseRoles parses role assignment rules (g type)
func parseRoles(policyStr string) [][]string {
	var roles [][]string
	lines := splitLines(policyStr)

	for _, line := range lines {
		line = trimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		parts := splitComma(line)
		if len(parts) > 0 && parts[0] == "g" {
			// Skip the policy type prefix
			parts = parts[1:]
			if len(parts) > 0 {
				roles = append(roles, parts)
			}
		}
	}

	return roles
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

func splitComma(s string) []string {
	var parts []string
	start := 0

	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			part := trimSpace(s[start:i])
			if part != "" {
				parts = append(parts, part)
			}
			start = i + 1
		}
	}

	if start < len(s) {
		part := trimSpace(s[start:])
		if part != "" {
			parts = append(parts, part)
		}
	}

	return parts
}

// HealthCheck handles health check requests
func (s *WebhookServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
