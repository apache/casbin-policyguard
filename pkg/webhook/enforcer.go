package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	policyv1alpha1 "github.com/casbin/policywall/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PolicyEnforcer handles policy enforcement with Casbin
type PolicyEnforcer struct {
	client   client.Client
	enforcer *casbin.Enforcer
}

// EnforcementResult contains the result of policy enforcement
type EnforcementResult struct {
	Allowed     bool
	Reason      string
	Patches     []PatchOperation
	Validations []ValidationResult
}

// PatchOperation represents a JSON Patch operation
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	Allowed bool
	Message string
}

// NewPolicyEnforcer creates a new policy enforcer
func NewPolicyEnforcer(client client.Client) (*PolicyEnforcer, error) {
	// Default Casbin model for RBAC with resource-based control
	defaultModel := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(defaultModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin model: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	return &PolicyEnforcer{
		client:   client,
		enforcer: enforcer,
	}, nil
}

// LoadPolicies loads policies from the cluster into Casbin
func (pe *PolicyEnforcer) LoadPolicies(ctx context.Context) error {
	logger := log.FromContext(ctx)

	var policyList policyv1alpha1.PolicyList
	if err := pe.client.List(ctx, &policyList); err != nil {
		return fmt.Errorf("failed to list policies: %w", err)
	}

	// Clear existing policies
	pe.enforcer.ClearPolicy()

	// Load policies into Casbin
	for _, policy := range policyList.Items {
		for _, subject := range policy.Spec.Subjects {
			for _, resource := range policy.Spec.Resources {
				for _, res := range resource.Resources {
					// Add validation rules as policies
					for _, rule := range policy.Spec.ValidationRules {
						if rule.Action == "allow" {
							_, err := pe.enforcer.AddPolicy(subject, res, "validate")
							if err != nil {
								logger.Error(err, "failed to add policy", "subject", subject, "resource", res)
							}
						}
					}
					// Add mutation rules as policies
					if len(policy.Spec.MutationRules) > 0 {
						_, err := pe.enforcer.AddPolicy(subject, res, "mutate")
						if err != nil {
							logger.Error(err, "failed to add policy", "subject", subject, "resource", res)
						}
					}
				}
			}
		}
	}

	return nil
}

// Enforce evaluates policies and returns patches and validation results
func (pe *PolicyEnforcer) Enforce(ctx context.Context, obj runtime.Object, username string, operation string) (*EnforcementResult, error) {
	logger := log.FromContext(ctx)

	result := &EnforcementResult{
		Allowed:     true,
		Patches:     []PatchOperation{},
		Validations: []ValidationResult{},
	}

	// Get resource type
	gvk := obj.GetObjectKind().GroupVersionKind()
	resourceType := strings.ToLower(gvk.Kind) + "s"

	// Load all policies
	var policyList policyv1alpha1.PolicyList
	if err := pe.client.List(ctx, &policyList); err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	// Convert object to unstructured for easier field access
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to convert object to unstructured: %w", err)
	}
	u := &unstructured.Unstructured{Object: unstructuredObj}

	// Process each policy
	for _, policy := range policyList.Items {
		// Check if policy applies to this resource
		if !pe.policyAppliesTo(&policy, resourceType, u) {
			continue
		}

		// Check if user is in subjects
		if !pe.subjectMatches(&policy, username) {
			continue
		}

		// Process validation rules
		for _, rule := range policy.Spec.ValidationRules {
			if pe.conditionsMatch(u, rule.Conditions) {
				validation := ValidationResult{
					Allowed: rule.Action == "allow",
					Message: rule.Message,
				}
				result.Validations = append(result.Validations, validation)

				if !validation.Allowed {
					result.Allowed = false
					result.Reason = rule.Message
					logger.Info("validation failed", "rule", rule.Name, "reason", rule.Message)
				}
			}
		}

		// Process mutation rules (only if operation allows mutation)
		if operation == "CREATE" || operation == "UPDATE" {
			for _, rule := range policy.Spec.MutationRules {
				if pe.conditionsMatch(u, rule.Conditions) {
					patch := pe.generatePatch(&rule)
					result.Patches = append(result.Patches, patch)
					logger.Info("mutation applied", "rule", rule.Name, "operation", rule.Operation, "path", rule.Path)
				}
			}
		}
	}

	return result, nil
}

// policyAppliesTo checks if a policy applies to the given resource
func (pe *PolicyEnforcer) policyAppliesTo(policy *policyv1alpha1.Policy, resourceType string, obj *unstructured.Unstructured) bool {
	namespace := obj.GetNamespace()

	for _, rs := range policy.Spec.Resources {
		// Check if resource type matches
		resourceMatches := false
		for _, res := range rs.Resources {
			if strings.ToLower(res) == strings.ToLower(resourceType) || res == "*" {
				resourceMatches = true
				break
			}
		}
		if !resourceMatches {
			continue
		}

		// Check namespace if specified
		if len(rs.Namespaces) > 0 {
			namespaceMatches := false
			for _, ns := range rs.Namespaces {
				if ns == namespace || ns == "*" {
					namespaceMatches = true
					break
				}
			}
			if !namespaceMatches {
				continue
			}
		}

		return true
	}

	return false
}

// subjectMatches checks if the username matches policy subjects
func (pe *PolicyEnforcer) subjectMatches(policy *policyv1alpha1.Policy, username string) bool {
	if len(policy.Spec.Subjects) == 0 {
		return true // No subjects means applies to all
	}

	for _, subject := range policy.Spec.Subjects {
		if subject == username || subject == "*" {
			return true
		}
	}

	return false
}

// conditionsMatch checks if all conditions are satisfied
func (pe *PolicyEnforcer) conditionsMatch(obj *unstructured.Unstructured, conditions []policyv1alpha1.RuleCondition) bool {
	if len(conditions) == 0 {
		return true // No conditions means always match
	}

	for _, cond := range conditions {
		if !pe.evaluateCondition(obj, &cond) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition
func (pe *PolicyEnforcer) evaluateCondition(obj *unstructured.Unstructured, cond *policyv1alpha1.RuleCondition) bool {
	// Get field value from object
	fieldValue, found, err := unstructured.NestedString(obj.Object, strings.Split(cond.Field, ".")...)
	if err != nil {
		return false
	}

	switch cond.Operator {
	case "exists":
		return found
	case "notExists":
		return !found
	case "equals":
		return found && fieldValue == cond.Value
	case "notEquals":
		return found && fieldValue != cond.Value
	case "in":
		if !found {
			return false
		}
		for _, v := range cond.Values {
			if fieldValue == v {
				return true
			}
		}
		return false
	case "notIn":
		if !found {
			return true
		}
		for _, v := range cond.Values {
			if fieldValue == v {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// generatePatch creates a JSON Patch operation from a mutation rule
func (pe *PolicyEnforcer) generatePatch(rule *policyv1alpha1.MutationRule) PatchOperation {
	patch := PatchOperation{
		Op:   rule.Operation,
		Path: rule.Path,
	}

	// Parse value as JSON if possible, otherwise use as string
	if rule.Value != "" {
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(rule.Value), &jsonValue); err == nil {
			patch.Value = jsonValue
		} else {
			patch.Value = rule.Value
		}
	}

	return patch
}

// Common mutation helpers

// GenerateSidecarInjectionPatch generates a patch to inject a sidecar container
func GenerateSidecarInjectionPatch(containerName, image string) []PatchOperation {
	sidecarContainer := corev1.Container{
		Name:  containerName,
		Image: image,
	}

	containerJSON, _ := json.Marshal(sidecarContainer)
	var containerMap map[string]interface{}
	json.Unmarshal(containerJSON, &containerMap)

	return []PatchOperation{
		{
			Op:    "add",
			Path:  "/spec/containers/-",
			Value: containerMap,
		},
	}
}

// GenerateResourceLimitsPatch generates a patch to set resource limits
func GenerateResourceLimitsPatch(cpu, memory string) []PatchOperation {
	return []PatchOperation{
		{
			Op:   "add",
			Path: "/spec/containers/0/resources/limits",
			Value: map[string]interface{}{
				"cpu":    cpu,
				"memory": memory,
			},
		},
	}
}

// GenerateLabelPatch generates a patch to add/update labels
func GenerateLabelPatch(labels map[string]string) []PatchOperation {
	patches := []PatchOperation{}
	for k, v := range labels {
		patches = append(patches, PatchOperation{
			Op:    "add",
			Path:  "/metadata/labels/" + strings.ReplaceAll(k, "/", "~1"),
			Value: v,
		})
	}
	return patches
}
