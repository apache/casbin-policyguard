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
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewCollector(t *testing.T) {
	registry := prometheus.NewRegistry()
	collector := NewCollectorWithRegistry(registry)
	if collector == nil {
		t.Fatal("expected non-nil collector")
	}
	if collector.admissionRequests == nil {
		t.Fatal("expected non-nil admissionRequests")
	}
	if collector.policyEvaluations == nil {
		t.Fatal("expected non-nil policyEvaluations")
	}
}

func TestRecordAdmissionRequest(t *testing.T) {
	registry := prometheus.NewRegistry()
	collector := NewCollectorWithRegistry(registry)

	// Should not panic
	collector.RecordAdmissionRequest("CREATE", true)
	collector.RecordAdmissionRequest("UPDATE", false)
}

func TestHandler(t *testing.T) {
	registry := prometheus.NewRegistry()
	collector := NewCollectorWithRegistry(registry)
	handler := collector.Handler()

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %d", w.Code)
	}
}
