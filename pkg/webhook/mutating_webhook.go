package webhook

import (
	"context"
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// MutatingWebhook handles mutating admission requests
type MutatingWebhook struct {
	client   client.Client
	enforcer *PolicyEnforcer
}

// NewMutatingWebhook creates a new mutating webhook handler
func NewMutatingWebhook(client client.Client, enforcer *PolicyEnforcer) *MutatingWebhook {
	return &MutatingWebhook{
		client:   client,
		enforcer: enforcer,
	}
}

// Handle processes mutating admission requests
func (m *MutatingWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)
	logger.Info("processing mutating webhook request",
		"operation", req.Operation,
		"resource", req.Resource.Resource,
		"name", req.Name,
		"namespace", req.Namespace,
		"username", req.UserInfo.Username,
	)

	// Decode the object as unstructured to support any resource type
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal(req.Object.Raw, obj); err != nil {
		logger.Error(err, "failed to decode object")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Enforce policies and get patches
	result, err := m.enforcer.Enforce(ctx, obj, req.UserInfo.Username, string(req.Operation))
	if err != nil {
		logger.Error(err, "failed to enforce policies")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// If there are no patches, allow the request as-is
	if len(result.Patches) == 0 {
		logger.Info("no mutations required")
		return admission.Allowed("no mutations required")
	}

	// Marshal patches to JSON
	patchBytes, err := json.Marshal(result.Patches)
	if err != nil {
		logger.Error(err, "failed to marshal patches")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	logger.Info("applying mutations", "patchCount", len(result.Patches))

	// Return a patched response
	return admission.PatchResponseFromRaw(req.Object.Raw, patchBytes)
}
