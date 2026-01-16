package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/policywall/api/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWebhookServer_DryRunMode(t *testing.T) {
	server := NewWebhookServer()

	// Create a policy with dry-run enabled
	policy := &v1alpha1.AdmissionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-policy",
		},
		Spec: v1alpha1.AdmissionPolicySpec{
			DryRun: true,
			Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
			Policy: `p, admin, production/test, DELETE`,
		},
	}

	err := server.UpdatePolicy(policy)
	if err != nil {
		t.Fatalf("Failed to update policy: %v", err)
	}

	// Create admission request that should violate the policy
	admissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "test-uid",
			Kind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			Resource: metav1.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			Name:      "test",
			Namespace: "production",
			Operation: admissionv1.Delete,
			UserInfo: authenticationv1.UserInfo{
				Username: "user", // Not admin
			},
		},
	}

	body, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response admissionv1.AdmissionReview
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// In dry-run mode, request should be allowed
	if !response.Response.Allowed {
		t.Errorf("Expected request to be allowed in dry-run mode, but it was denied")
	}

	// Should have warnings about the violation
	if len(response.Response.Warnings) == 0 {
		t.Errorf("Expected warnings in dry-run mode, but got none")
	}
}

func TestWebhookServer_EnforceMode(t *testing.T) {
	server := NewWebhookServer()

	// Create a policy with dry-run disabled (enforcement mode)
	policy := &v1alpha1.AdmissionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-policy",
		},
		Spec: v1alpha1.AdmissionPolicySpec{
			DryRun: false,
			Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
			Policy: `p, admin, production/test, DELETE`,
		},
	}

	err := server.UpdatePolicy(policy)
	if err != nil {
		t.Fatalf("Failed to update policy: %v", err)
	}

	// Create admission request that violates the policy
	admissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "test-uid",
			Kind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			Resource: metav1.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			Name:      "test",
			Namespace: "production",
			Operation: admissionv1.Delete,
			UserInfo: authenticationv1.UserInfo{
				Username: "user", // Not admin
			},
		},
	}

	body, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response admissionv1.AdmissionReview
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// In enforce mode, request should be denied
	if response.Response.Allowed {
		t.Errorf("Expected request to be denied in enforce mode, but it was allowed")
	}
}

func TestWebhookServer_AllowedRequest(t *testing.T) {
	server := NewWebhookServer()

	// Create a policy
	policy := &v1alpha1.AdmissionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-policy",
		},
		Spec: v1alpha1.AdmissionPolicySpec{
			DryRun: false,
			Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
			Policy: `p, admin, production/test, DELETE`,
		},
	}

	err := server.UpdatePolicy(policy)
	if err != nil {
		t.Fatalf("Failed to update policy: %v", err)
	}

	// Create admission request that satisfies the policy
	admissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "test-uid",
			Kind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			Resource: metav1.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			Name:      "test",
			Namespace: "production",
			Operation: admissionv1.Delete,
			UserInfo: authenticationv1.UserInfo{
				Username: "admin", // Matches policy
			},
		},
	}

	body, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response admissionv1.AdmissionReview
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Request should be allowed
	if !response.Response.Allowed {
		t.Errorf("Expected request to be allowed, but it was denied")
	}

	// Should have no warnings
	if len(response.Response.Warnings) > 0 {
		t.Errorf("Expected no warnings for allowed request, but got %d", len(response.Response.Warnings))
	}
}

func TestWebhookServer_MatchRules(t *testing.T) {
	server := NewWebhookServer()

	// Create a policy with match rules
	policy := &v1alpha1.AdmissionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-policy",
		},
		Spec: v1alpha1.AdmissionPolicySpec{
			DryRun: false,
			Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
			Policy: `p, admin, default/test, DELETE`,
			MatchRules: []v1alpha1.MatchRule{
				{
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Resources:   []string{"pods"},
					Operations:  []string{"DELETE"},
				},
			},
		},
	}

	err := server.UpdatePolicy(policy)
	if err != nil {
		t.Fatalf("Failed to update policy: %v", err)
	}

	// Request that matches the rules
	admissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "test-uid",
			Kind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			Resource: metav1.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			Name:      "test",
			Namespace: "default",
			Operation: admissionv1.Delete,
			UserInfo: authenticationv1.UserInfo{
				Username: "user",
			},
		},
	}

	body, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should be denied because user != admin
	if response.Response.Allowed {
		t.Errorf("Expected request to be denied")
	}
}

func TestWebhookServer_MultiplePolices(t *testing.T) {
	server := NewWebhookServer()

	// Add a dry-run policy
	policy1 := &v1alpha1.AdmissionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dryrun-policy",
		},
		Spec: v1alpha1.AdmissionPolicySpec{
			DryRun: true,
			Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
			Policy: `p, admin, default/test, CREATE`,
		},
	}

	// Add an enforce policy
	policy2 := &v1alpha1.AdmissionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "enforce-policy",
		},
		Spec: v1alpha1.AdmissionPolicySpec{
			DryRun: false,
			Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
			Policy: `p, superadmin, default/test, CREATE`,
		},
	}

	err := server.UpdatePolicy(policy1)
	if err != nil {
		t.Fatalf("Failed to update policy1: %v", err)
	}

	err = server.UpdatePolicy(policy2)
	if err != nil {
		t.Fatalf("Failed to update policy2: %v", err)
	}

	// Request that violates both policies
	admissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Request: &admissionv1.AdmissionRequest{
			UID: "test-uid",
			Kind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Pod",
			},
			Resource: metav1.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: "pods",
			},
			Name:      "test",
			Namespace: "default",
			Operation: admissionv1.Create,
			UserInfo: authenticationv1.UserInfo{
				Username: "user",
			},
		},
	}

	body, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should be denied because of enforce-policy
	if response.Response.Allowed {
		t.Errorf("Expected request to be denied by enforce-policy")
	}
}

func TestHealthCheck(t *testing.T) {
	server := NewWebhookServer()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	server.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}
