package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicySpec defines the desired state of Policy
type PolicySpec struct {
	// ValidationRules defines rules for validating resources
	// +optional
	ValidationRules []ValidationRule `json:"validationRules,omitempty"`

	// MutationRules defines rules for mutating resources
	// +optional
	MutationRules []MutationRule `json:"mutationRules,omitempty"`

	// Subjects defines who can perform actions (users, service accounts, groups)
	// +optional
	Subjects []string `json:"subjects,omitempty"`

	// Resources defines which Kubernetes resources this policy applies to
	Resources []ResourceSelector `json:"resources"`

	// CasbinModel defines the Casbin model configuration
	// +optional
	CasbinModel string `json:"casbinModel,omitempty"`
}

// ValidationRule defines a single validation rule
type ValidationRule struct {
	// Name of the validation rule
	Name string `json:"name"`

	// Action defines what action to take (allow, deny)
	Action string `json:"action"`

	// Conditions defines conditions that must be met
	// +optional
	Conditions []RuleCondition `json:"conditions,omitempty"`

	// Message is the message to return when validation fails
	// +optional
	Message string `json:"message,omitempty"`
}

// MutationRule defines a single mutation rule
type MutationRule struct {
	// Name of the mutation rule
	Name string `json:"name"`

	// Operation defines the JSON Patch operation (add, remove, replace)
	Operation string `json:"operation"`

	// Path defines the JSON path to mutate
	Path string `json:"path"`

	// Value defines the value to set (for add/replace operations)
	// +optional
	Value string `json:"value,omitempty"`

	// Priority determines the order of patch application (lower values = higher priority)
	// Default is 0. Patches are sorted by priority in ascending order.
	// +optional
	Priority int `json:"priority,omitempty"`

	// Template references a predefined mutation template (e.g., "sidecar", "resource-limits")
	// When set, the helper function generates the patch value
	// +optional
	Template string `json:"template,omitempty"`

	// TemplateParams provides parameters for the template
	// +optional
	TemplateParams map[string]string `json:"templateParams,omitempty"`

	// Conditions defines when this mutation should apply
	// +optional
	Conditions []RuleCondition `json:"conditions,omitempty"`
}

// RuleCondition defines a condition for rule matching
type RuleCondition struct {
	// Field is the JSON path to check
	Field string `json:"field"`

	// Operator defines the comparison operator (equals, notEquals, in, notIn, exists, notExists)
	Operator string `json:"operator"`

	// Value is the value to compare against
	// +optional
	Value string `json:"value,omitempty"`

	// Values is used for 'in' and 'notIn' operators
	// +optional
	Values []string `json:"values,omitempty"`
}

// ResourceSelector defines which resources a policy applies to
type ResourceSelector struct {
	// APIGroups is a list of API groups
	// +optional
	APIGroups []string `json:"apiGroups,omitempty"`

	// APIVersions is a list of API versions
	// +optional
	APIVersions []string `json:"apiVersions,omitempty"`

	// Resources is a list of resource types (e.g., "pods", "deployments")
	Resources []string `json:"resources"`

	// Namespaces is a list of namespaces to apply the policy to
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`
}

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {
	// Conditions represent the latest available observations of the policy's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastUpdateTime is the last time the policy was updated
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// AppliedCount is the number of times this policy has been successfully applied
	// +optional
	AppliedCount int64 `json:"appliedCount,omitempty"`

	// RejectedCount is the number of times this policy has rejected a request
	// +optional
	RejectedCount int64 `json:"rejectedCount,omitempty"`

	// LastAppliedTime is the timestamp of the last successful application
	// +optional
	LastAppliedTime metav1.Time `json:"lastAppliedTime,omitempty"`

	// ErrorMessage contains the last error message if policy application failed
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Policy is the Schema for the policies API
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec,omitempty"`
	Status PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
}
