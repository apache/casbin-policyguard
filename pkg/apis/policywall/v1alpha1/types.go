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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AdmissionPolicy defines an admission control policy
type AdmissionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdmissionPolicySpec   `json:"spec,omitempty"`
	Status AdmissionPolicyStatus `json:"status,omitempty"`
}

// AdmissionPolicySpec defines the desired state of AdmissionPolicy
type AdmissionPolicySpec struct {
	// Template specifies the type of policy template to use
	Template string `json:"template"`

	// Model defines the Casbin RBAC model
	Model string `json:"model"`

	// Rules contains the policy rules
	Rules []string `json:"rules"`

	// FailurePolicy defines how to handle policy evaluation failures
	// +optional
	FailurePolicy string `json:"failurePolicy,omitempty"`

	// NamespaceSelector determines which namespaces this policy applies to
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// DryRun when true, policy violations are logged but not enforced
	// +optional
	DryRun bool `json:"dryRun,omitempty"`

	// Parameters contains template-specific parameters
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`
}

// AdmissionPolicyStatus defines the observed state of AdmissionPolicy
type AdmissionPolicyStatus struct {
	// Phase indicates the current phase of the policy
	Phase string `json:"phase,omitempty"`

	// Message provides human-readable information about the status
	// +optional
	Message string `json:"message,omitempty"`

	// LastUpdated is the timestamp of the last status update
	// +optional
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`

	// ViolationCount tracks the number of policy violations
	// +optional
	ViolationCount int64 `json:"violationCount,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AdmissionPolicyList contains a list of AdmissionPolicy
type AdmissionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdmissionPolicy `json:"items"`
}
