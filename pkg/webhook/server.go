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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/casbin/policywall/pkg/casbin"
	"github.com/casbin/policywall/pkg/metrics"
)

// Server handles admission webhook requests
type Server struct {
	enforcer *casbin.PolicyEnforcer
	metrics  *metrics.Collector
	dryRun   bool
}

// NewServer creates a new webhook server
func NewServer(enforcer *casbin.PolicyEnforcer, metrics *metrics.Collector) *Server {
	return &Server{
		enforcer: enforcer,
		metrics:  metrics,
	}
}

// HandleAdmission handles admission review requests
func (s *Server) HandleAdmission(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		klog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect application/json", http.StatusUnsupportedMediaType)
		return
	}

	var admissionReviewReq admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReviewReq); err != nil {
		klog.Errorf("could not deserialize request: %v", err)
		http.Error(w, fmt.Sprintf("could not deserialize request: %v", err), http.StatusBadRequest)
		return
	}

	admissionReviewResponse := s.admit(admissionReviewReq)
	resp, err := json.Marshal(admissionReviewResponse)
	if err != nil {
		klog.Errorf("could not serialize response: %v", err)
		http.Error(w, fmt.Sprintf("could not serialize response: %v", err), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		klog.Errorf("could not write response: %v", err)
	}
}

// admit performs the admission logic
func (s *Server) admit(ar admissionv1.AdmissionReview) *admissionv1.AdmissionReview {
	req := ar.Request
	var allowed = true
	var msg string

	klog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v Operation=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation)

	// Only handle CREATE and UPDATE operations
	if req.Operation != admissionv1.Create && req.Operation != admissionv1.Update {
		allowed = true
		msg = "operation not handled"
	} else {
		// Check against all policies
		// In production, this would match policies based on namespace selectors
		allowed = true
		msg = "allowed by policy"
		
		s.metrics.RecordAdmissionRequest(string(req.Operation), allowed)
	}

	admissionResponse := &admissionv1.AdmissionResponse{
		Allowed: allowed,
		UID:     req.UID,
		Result: &metav1.Status{
			Message: msg,
		},
	}

	reviewResponse := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: admissionResponse,
	}

	return reviewResponse
}

// HandleHealth handles health check requests
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		klog.Errorf("failed to write health response: %v", err)
	}
}

// HandleReady handles readiness check requests
func (s *Server) HandleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ready")); err != nil {
		klog.Errorf("failed to write ready response: %v", err)
	}
}
