package controller

import (
	"context"
	"fmt"
	"sync"
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
	finalizerName        = "policy.casbin.org/finalizer"
	maxRecentViolations  = 10
	statusUpdateInterval = 30 * time.Second
)

// AdmissionPolicyReconciler reconciles AdmissionPolicy objects
type AdmissionPolicyReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	WebhookServer *webhook.WebhookServer

	// Track violations per policy
	mu         sync.RWMutex
	violations map[string][]v1alpha1.ViolationResource

	// Last status update time per policy
	lastStatusUpdate map[string]time.Time
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

	// Update status to reflect success with recent violations
	r.mu.RLock()
	recentViolations := r.violations[policy.Name]
	violationCount := len(recentViolations)
	r.mu.RUnlock()

	policy.Status.Ready = true
	policy.Status.Message = "Policy loaded successfully"
	if policy.Spec.DryRun {
		policy.Status.Message = fmt.Sprintf("Policy loaded in dry-run mode (audit only). %d violations detected.", violationCount)
	}
	policy.Status.LastUpdated = metav1.Now()
	policy.Status.ViolationCount = int64(violationCount)

	// Include recent violations (last 10)
	if len(recentViolations) > 0 {
		startIdx := 0
		if len(recentViolations) > maxRecentViolations {
			startIdx = len(recentViolations) - maxRecentViolations
		}
		policy.Status.RecentViolations = recentViolations[startIdx:]
	} else {
		policy.Status.RecentViolations = nil
	}

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

// RecordViolation handles violation reporting from the webhook
func (r *AdmissionPolicyReconciler) RecordViolation(policyName string, violation v1alpha1.ViolationResource) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.violations == nil {
		r.violations = make(map[string][]v1alpha1.ViolationResource)
	}

	// Add the violation
	r.violations[policyName] = append(r.violations[policyName], violation)

	// Keep only the most recent violations (up to 100 to prevent unbounded growth)
	if len(r.violations[policyName]) > 100 {
		r.violations[policyName] = r.violations[policyName][len(r.violations[policyName])-100:]
	}
}

// SetupWithManager sets up the controller with the Manager
func (r *AdmissionPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize maps
	if r.violations == nil {
		r.violations = make(map[string][]v1alpha1.ViolationResource)
	}
	if r.lastStatusUpdate == nil {
		r.lastStatusUpdate = make(map[string]time.Time)
	}

	// Set up violation callback
	r.WebhookServer.SetViolationCallback(r.RecordViolation)

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AdmissionPolicy{}).
		Complete(r)
}
