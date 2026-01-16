package webhook

import (
	"context"
	"testing"

	policyv1alpha1 "github.com/casbin/policywall/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPolicyEnforcer_Enforce_MutationRules(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = policyv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create a policy with mutation rules
	policy := &policyv1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-mutation-policy",
		},
		Spec: policyv1alpha1.PolicySpec{
			Subjects: []string{"*"},
			Resources: []policyv1alpha1.ResourceSelector{
				{
					Resources: []string{"pods"},
				},
			},
			MutationRules: []policyv1alpha1.MutationRule{
				{
					Name:      "add-label",
					Operation: "add",
					Path:      "/metadata/labels/mutated",
					Value:     `"true"`,
				},
			},
		},
	}

	// Create fake client with policy
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(policy).
		Build()

	enforcer, err := NewPolicyEnforcer(fakeClient)
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	// Create a pod to test
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "nginx",
				},
			},
		},
	}

	ctx := context.Background()
	result, err := enforcer.Enforce(ctx, pod, "test-user", "CREATE")
	if err != nil {
		t.Fatalf("enforce failed: %v", err)
	}

	if !result.Allowed {
		t.Errorf("expected allowed=true, got false")
	}

	if len(result.Patches) != 1 {
		t.Errorf("expected 1 patch, got %d", len(result.Patches))
	}

	if len(result.Patches) > 0 {
		patch := result.Patches[0]
		if patch.Op != "add" {
			t.Errorf("expected patch op=add, got %s", patch.Op)
		}
		if patch.Path != "/metadata/labels/mutated" {
			t.Errorf("expected path=/metadata/labels/mutated, got %s", patch.Path)
		}
	}
}

func TestPolicyEnforcer_Enforce_ValidationRules(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = policyv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create a policy with validation rules
	policy := &policyv1alpha1.Policy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-validation-policy",
		},
		Spec: policyv1alpha1.PolicySpec{
			Subjects: []string{"*"},
			Resources: []policyv1alpha1.ResourceSelector{
				{
					Resources: []string{"pods"},
				},
			},
			ValidationRules: []policyv1alpha1.ValidationRule{
				{
					Name:    "require-label",
					Action:  "deny",
					Message: "Pod must have required-label",
					Conditions: []policyv1alpha1.RuleCondition{
						{
							Field:    "metadata.labels.required-label",
							Operator: "notExists",
						},
					},
				},
			},
		},
	}

	// Create fake client with policy
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(policy).
		Build()

	enforcer, err := NewPolicyEnforcer(fakeClient)
	if err != nil {
		t.Fatalf("failed to create enforcer: %v", err)
	}

	// Create a pod without the required label
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "nginx",
				},
			},
		},
	}

	ctx := context.Background()
	result, err := enforcer.Enforce(ctx, pod, "test-user", "CREATE")
	if err != nil {
		t.Fatalf("enforce failed: %v", err)
	}

	if result.Allowed {
		t.Errorf("expected allowed=false, got true")
	}

	if result.Reason != "Pod must have required-label" {
		t.Errorf("expected reason='Pod must have required-label', got '%s'", result.Reason)
	}
}

func TestGenerateSidecarInjectionPatch(t *testing.T) {
	patches := GenerateSidecarInjectionPatch("sidecar", "envoy:latest")

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]
	if patch.Op != "add" {
		t.Errorf("expected op=add, got %s", patch.Op)
	}
	if patch.Path != "/spec/containers/-" {
		t.Errorf("expected path=/spec/containers/-, got %s", patch.Path)
	}

	// Verify the container structure
	containerMap, ok := patch.Value.(map[string]interface{})
	if !ok {
		t.Fatal("expected container value to be a map")
	}
	if containerMap["name"] != "sidecar" {
		t.Errorf("expected container name=sidecar, got %v", containerMap["name"])
	}
	if containerMap["image"] != "envoy:latest" {
		t.Errorf("expected container image=envoy:latest, got %v", containerMap["image"])
	}
}

func TestGenerateResourceLimitsPatch(t *testing.T) {
	patches := GenerateResourceLimitsPatch("500m", "512Mi")

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch := patches[0]
	if patch.Op != "add" {
		t.Errorf("expected op=add, got %s", patch.Op)
	}
	if patch.Path != "/spec/containers/0/resources/limits" {
		t.Errorf("expected path=/spec/containers/0/resources/limits, got %s", patch.Path)
	}
}

func TestGenerateLabelPatch(t *testing.T) {
	labels := map[string]string{
		"app":     "test",
		"version": "v1",
	}
	patches := GenerateLabelPatch(labels)

	if len(patches) != 2 {
		t.Fatalf("expected 2 patches, got %d", len(patches))
	}

	for _, patch := range patches {
		if patch.Op != "add" {
			t.Errorf("expected op=add, got %s", patch.Op)
		}
		if patch.Value != "test" && patch.Value != "v1" {
			t.Errorf("unexpected patch value: %v", patch.Value)
		}
	}
}
