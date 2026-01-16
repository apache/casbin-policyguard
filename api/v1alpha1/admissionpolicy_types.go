package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdmissionPolicySpec defines the desired state of AdmissionPolicy
type AdmissionPolicySpec struct {
	// DryRun enables audit mode where violations are logged but requests are allowed
	// +optional
	DryRun bool `json:"dryRun,omitempty"`

	// Model is the Casbin model configuration
	Model string `json:"model"`

	// Policy is the Casbin policy configuration
	Policy string `json:"policy"`

	// MatchRules defines which admission requests this policy applies to
	// +optional
	MatchRules []MatchRule `json:"matchRules,omitempty"`
}

// MatchRule defines matching criteria for admission requests
type MatchRule struct {
	// APIGroups is the API groups the rule applies to
	// +optional
	APIGroups []string `json:"apiGroups,omitempty"`

	// APIVersions is the API versions the rule applies to
	// +optional
	APIVersions []string `json:"apiVersions,omitempty"`

	// Resources is the resources the rule applies to
	// +optional
	Resources []string `json:"resources,omitempty"`

	// Operations is the operations the rule applies to (CREATE, UPDATE, DELETE, CONNECT)
	// +optional
	Operations []string `json:"operations,omitempty"`
}

// ViolationResource represents a resource that violated a policy
type ViolationResource struct {
	// Kind is the resource kind (e.g., Pod, Deployment)
	Kind string `json:"kind"`

	// Namespace is the resource namespace
	Namespace string `json:"namespace"`

	// Name is the resource name
	Name string `json:"name"`

	// Operation is the operation that was attempted (CREATE, UPDATE, DELETE)
	Operation string `json:"operation"`

	// Timestamp is when the violation occurred
	Timestamp metav1.Time `json:"timestamp"`

	// User is the user who attempted the operation
	User string `json:"user,omitempty"`

	// Message is the violation reason
	Message string `json:"message,omitempty"`
}

// AdmissionPolicyStatus defines the observed state of AdmissionPolicy
type AdmissionPolicyStatus struct {
	// Ready indicates whether the policy is loaded and ready
	Ready bool `json:"ready"`

	// Message provides additional information about the status
	// +optional
	Message string `json:"message,omitempty"`

	// LastUpdated is the timestamp when the policy was last updated
	// +optional
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`

	// ViolationCount tracks the number of violations detected (in dry-run mode)
	// +optional
	ViolationCount int64 `json:"violationCount,omitempty"`

	// RecentViolations contains the most recent violations (limited to last 10)
	// +optional
	RecentViolations []ViolationResource `json:"recentViolations,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="DryRun",type=boolean,JSONPath=`.spec.dryRun`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Violations",type=integer,JSONPath=`.status.violationCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AdmissionPolicy is the Schema for the admissionpolicies API
type AdmissionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdmissionPolicySpec   `json:"spec,omitempty"`
	Status AdmissionPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AdmissionPolicyList contains a list of AdmissionPolicy
type AdmissionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdmissionPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AdmissionPolicy{}, &AdmissionPolicyList{})
}
