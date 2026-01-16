package webhook

import (
	"context"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ValidatingWebhook handles validating admission requests
type ValidatingWebhook struct {
	client   client.Client
	decoder  admission.Decoder
	enforcer *PolicyEnforcer
}

// NewValidatingWebhook creates a new validating webhook handler
func NewValidatingWebhook(client client.Client, enforcer *PolicyEnforcer) *ValidatingWebhook {
	return &ValidatingWebhook{
		client:   client,
		enforcer: enforcer,
	}
}

// Handle processes validating admission requests
func (v *ValidatingWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)
	logger.Info("processing validating webhook request",
		"operation", req.Operation,
		"resource", req.Resource.Resource,
		"name", req.Name,
		"namespace", req.Namespace,
		"username", req.UserInfo.Username,
	)

	// Decode the object
	obj := &corev1.Pod{}
	err := v.decoder.Decode(req, obj)
	if err != nil {
		logger.Error(err, "failed to decode object")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Enforce policies
	result, err := v.enforcer.Enforce(ctx, obj, req.UserInfo.Username, string(req.Operation))
	if err != nil {
		logger.Error(err, "failed to enforce policies")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Check validation results
	if !result.Allowed {
		logger.Info("validation failed", "reason", result.Reason)
		return admission.Denied(result.Reason)
	}

	// Check individual validations
	for _, validation := range result.Validations {
		if !validation.Allowed {
			logger.Info("validation rule failed", "message", validation.Message)
			return admission.Denied(validation.Message)
		}
	}

	logger.Info("validation passed")
	return admission.Allowed("validation passed")
}

// InjectDecoder injects the decoder
func (v *ValidatingWebhook) InjectDecoder(d admission.Decoder) error {
	v.decoder = d
	return nil
}
