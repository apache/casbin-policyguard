package webhook

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/policywall/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// AdmissionPolicyValidator validates AdmissionPolicy resources
type AdmissionPolicyValidator struct{}

// ValidateCreate validates creation of AdmissionPolicy
func (v *AdmissionPolicyValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	policy, ok := obj.(*v1alpha1.AdmissionPolicy)
	if !ok {
		return nil, fmt.Errorf("expected AdmissionPolicy but got %T", obj)
	}

	return v.validatePolicy(policy)
}

// ValidateUpdate validates updates to AdmissionPolicy
func (v *AdmissionPolicyValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	policy, ok := newObj.(*v1alpha1.AdmissionPolicy)
	if !ok {
		return nil, fmt.Errorf("expected AdmissionPolicy but got %T", newObj)
	}

	return v.validatePolicy(policy)
}

// ValidateDelete validates deletion of AdmissionPolicy
func (v *AdmissionPolicyValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No validation needed for deletion
	return nil, nil
}

// validatePolicy performs validation on AdmissionPolicy
func (v *AdmissionPolicyValidator) validatePolicy(policy *v1alpha1.AdmissionPolicy) (admission.Warnings, error) {
	var warnings admission.Warnings

	// Validate Casbin model
	if policy.Spec.Model == "" {
		return nil, fmt.Errorf("spec.model is required")
	}

	_, err := model.NewModelFromString(policy.Spec.Model)
	if err != nil {
		return nil, fmt.Errorf("invalid Casbin model: %v", err)
	}

	// Validate policy is not empty
	if policy.Spec.Policy == "" {
		warnings = append(warnings, "spec.policy is empty - this policy will deny all requests")
	}

	// Validate match rules if specified
	for i, rule := range policy.Spec.MatchRules {
		if len(rule.APIGroups) == 0 && len(rule.APIVersions) == 0 &&
			len(rule.Resources) == 0 && len(rule.Operations) == 0 {
			warnings = append(warnings, fmt.Sprintf("matchRules[%d] has no filters - it will match all requests", i))
		}
	}

	// Warn if dry-run is disabled for a new policy
	if !policy.Spec.DryRun && policy.CreationTimestamp.IsZero() {
		warnings = append(warnings, "Creating policy with dryRun=false. Consider testing with dryRun=true first.")
	}

	return warnings, nil
}
