package webhook

import (
	"context"
	"testing"

	"github.com/casbin/policywall/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAdmissionPolicyValidator_ValidateCreate(t *testing.T) {
	validator := &AdmissionPolicyValidator{}

	tests := []struct {
		name         string
		policy       *v1alpha1.AdmissionPolicy
		wantErr      bool
		wantWarnings int
	}{
		{
			name: "valid policy",
			policy: &v1alpha1.AdmissionPolicy{
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
m = r.sub == p.sub`,
					Policy: `p, admin, /resource, READ`,
				},
			},
			wantErr:      false,
			wantWarnings: 0,
		},
		{
			name: "invalid model",
			policy: &v1alpha1.AdmissionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy",
				},
				Spec: v1alpha1.AdmissionPolicySpec{
					Model:  `invalid model syntax`,
					Policy: `p, admin`,
				},
			},
			wantErr: true,
		},
		{
			name: "empty model",
			policy: &v1alpha1.AdmissionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-policy",
				},
				Spec: v1alpha1.AdmissionPolicySpec{
					Model:  "",
					Policy: `p, admin`,
				},
			},
			wantErr: true,
		},
		{
			name: "empty policy with warning",
			policy: &v1alpha1.AdmissionPolicy{
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
m = r.sub == p.sub`,
					Policy: "",
				},
			},
			wantErr:      false,
			wantWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := validator.ValidateCreate(context.Background(), tt.policy)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(warnings) != tt.wantWarnings {
				t.Errorf("ValidateCreate() got %d warnings, want %d", len(warnings), tt.wantWarnings)
			}
		})
	}
}
