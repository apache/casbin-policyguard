package controller

import (
	"context"
	"time"

	"github.com/casbin/policywall/api/v1alpha1"
	"github.com/casbin/policywall/pkg/webhook"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	finalizerName = "policy.casbin.org/finalizer"
)

// AdmissionPolicyReconciler reconciles AdmissionPolicy objects
type AdmissionPolicyReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	WebhookServer *webhook.WebhookServer
}

// Reconcile handles AdmissionPolicy changes
func (r *AdmissionPolicyReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	klog.Infof("Reconciling AdmissionPolicy %s", req.Name)

	// Fetch the AdmissionPolicy
	policy := &v1alpha1.AdmissionPolicy{}
	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		if client.IgnoreNotFound(err) == nil {
			// Policy was deleted
			klog.Infof("AdmissionPolicy %s not found, assuming deleted", req.Name)
			return reconcile.Result{}, nil
		}
		klog.Errorf("Failed to get AdmissionPolicy %s: %v", req.Name, err)
		return reconcile.Result{}, err
	}

	// Handle deletion
	if !policy.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, policy)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(policy, finalizerName) {
		controllerutil.AddFinalizer(policy, finalizerName)
		if err := r.Update(ctx, policy); err != nil {
			klog.Errorf("Failed to add finalizer to %s: %v", policy.Name, err)
			return reconcile.Result{}, err
		}
	}

	// Update webhook server with the policy
	if err := r.WebhookServer.UpdatePolicy(policy); err != nil {
		klog.Errorf("Failed to update webhook policy %s: %v", policy.Name, err)

		// Update status to reflect error
		policy.Status.Ready = false
		policy.Status.Message = err.Error()
		policy.Status.LastUpdated = metav1.Now()
		if err := r.Status().Update(ctx, policy); err != nil {
			klog.Errorf("Failed to update status for %s: %v", policy.Name, err)
		}

		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	// Update status to reflect success
	policy.Status.Ready = true
	policy.Status.Message = "Policy loaded successfully"
	if policy.Spec.DryRun {
		policy.Status.Message = "Policy loaded in dry-run mode (audit only)"
	}
	policy.Status.LastUpdated = metav1.Now()

	if err := r.Status().Update(ctx, policy); err != nil {
		klog.Errorf("Failed to update status for %s: %v", policy.Name, err)
		return reconcile.Result{}, err
	}

	klog.Infof("Successfully reconciled AdmissionPolicy %s (dryRun: %v)", policy.Name, policy.Spec.DryRun)
	return reconcile.Result{}, nil
}

// handleDeletion handles policy deletion
func (r *AdmissionPolicyReconciler) handleDeletion(ctx context.Context, policy *v1alpha1.AdmissionPolicy) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(policy, finalizerName) {
		return reconcile.Result{}, nil
	}

	// Remove policy from webhook server
	r.WebhookServer.DeletePolicy(policy.Name)

	// Remove finalizer
	controllerutil.RemoveFinalizer(policy, finalizerName)
	if err := r.Update(ctx, policy); err != nil {
		klog.Errorf("Failed to remove finalizer from %s: %v", policy.Name, err)
		return reconcile.Result{}, err
	}

	klog.Infof("Successfully deleted AdmissionPolicy %s", policy.Name)
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager
func (r *AdmissionPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AdmissionPolicy{}).
		Complete(r)
}
