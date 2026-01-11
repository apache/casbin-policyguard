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

package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/casbin/policywall/pkg/casbin"
	"github.com/casbin/policywall/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func TestHandleHealth(t *testing.T) {
	enforcer := casbin.NewPolicyEnforcer()
	registry := prometheus.NewRegistry()
	metricsCollector := metrics.NewCollectorWithRegistry(registry)
	server := NewServer(enforcer, metricsCollector)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	if w.Body.String() != "ok" {
		t.Errorf("expected 'ok', got %s", w.Body.String())
	}
}

func TestHandleReady(t *testing.T) {
	enforcer := casbin.NewPolicyEnforcer()
	registry := prometheus.NewRegistry()
	metricsCollector := metrics.NewCollectorWithRegistry(registry)
	server := NewServer(enforcer, metricsCollector)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	server.HandleReady(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	if w.Body.String() != "ready" {
		t.Errorf("expected 'ready', got %s", w.Body.String())
	}
}

func TestHandleAdmission(t *testing.T) {
	enforcer := casbin.NewPolicyEnforcer()
	registry := prometheus.NewRegistry()
	metricsCollector := metrics.NewCollectorWithRegistry(registry)
	server := NewServer(enforcer, metricsCollector)

	admissionReview := admissionv1.AdmissionReview{
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
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: []byte(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test"}}`),
			},
		},
	}

	body, err := json.Marshal(admissionReview)
	if err != nil {
		t.Fatalf("failed to marshal admission review: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleAdmission(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}

	var response admissionv1.AdmissionReview
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Response == nil {
		t.Fatal("expected non-nil response")
	}

	if response.Response.UID != "test-uid" {
		t.Errorf("expected UID 'test-uid', got %s", response.Response.UID)
	}
}

func TestHandleAdmissionEmptyBody(t *testing.T) {
	enforcer := casbin.NewPolicyEnforcer()
	registry := prometheus.NewRegistry()
	metricsCollector := metrics.NewCollectorWithRegistry(registry)
	server := NewServer(enforcer, metricsCollector)

	req := httptest.NewRequest(http.MethodPost, "/validate", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.HandleAdmission(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status BadRequest, got %d", w.Code)
	}
}

func TestHandleAdmissionInvalidContentType(t *testing.T) {
	enforcer := casbin.NewPolicyEnforcer()
	registry := prometheus.NewRegistry()
	metricsCollector := metrics.NewCollectorWithRegistry(registry)
	server := NewServer(enforcer, metricsCollector)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	server.HandleAdmission(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected status UnsupportedMediaType, got %d", w.Code)
	}
}
