// Copyright 2026 The Casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/casbin/policywall/pkg/apis/policywall/v1alpha1"
	"github.com/casbin/policywall/pkg/casbin"
)

// PolicyController watches AdmissionPolicy resources and updates the enforcer
type PolicyController struct {
	clientset kubernetes.Interface
	enforcer  *casbin.PolicyEnforcer
	informer  cache.SharedIndexInformer
	stopCh    chan struct{}
}

// NewPolicyController creates a new PolicyController
func NewPolicyController(clientset kubernetes.Interface, enforcer *casbin.PolicyEnforcer) *PolicyController {
	return &PolicyController{
		clientset: clientset,
		enforcer:  enforcer,
		stopCh:    make(chan struct{}),
	}
}

// Run starts the controller
func (pc *PolicyController) Run() error {
	klog.Info("Starting PolicyController")

	// Create informer for AdmissionPolicy resources
	pc.informer = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				// In production, this would use a typed client
				return &v1alpha1.AdmissionPolicyList{}, nil
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				// In production, this would use a typed client
				return nil, nil
			},
		},
		&v1alpha1.AdmissionPolicy{},
		time.Minute*10,
		cache.Indexers{},
	)

	pc.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pc.onAdd,
		UpdateFunc: pc.onUpdate,
		DeleteFunc: pc.onDelete,
	})

	go pc.informer.Run(pc.stopCh)

	if !cache.WaitForCacheSync(pc.stopCh, pc.informer.HasSynced) {
		return fmt.Errorf("failed to sync cache")
	}

	klog.Info("PolicyController started successfully")
	<-pc.stopCh
	return nil
}

// Stop stops the controller
func (pc *PolicyController) Stop() {
	klog.Info("Stopping PolicyController")
	close(pc.stopCh)
}

func (pc *PolicyController) onAdd(obj interface{}) {
	policy, ok := obj.(*v1alpha1.AdmissionPolicy)
	if !ok {
		klog.Errorf("unexpected object type: %T", obj)
		return
	}

	klog.Infof("Adding policy: %s/%s", policy.Namespace, policy.Name)
	
	modelText := policy.Spec.Model
	if policy.Spec.Template != "" {
		if tmpl, ok := casbin.GetTemplate(policy.Spec.Template); ok {
			modelText = tmpl.Model
		}
	}

	if err := pc.enforcer.AddPolicy(
		fmt.Sprintf("%s/%s", policy.Namespace, policy.Name),
		modelText,
		policy.Spec.Rules,
	); err != nil {
		klog.Errorf("failed to add policy: %v", err)
	}
}

func (pc *PolicyController) onUpdate(oldObj, newObj interface{}) {
	oldPolicy, ok := oldObj.(*v1alpha1.AdmissionPolicy)
	if !ok {
		klog.Errorf("unexpected object type: %T", oldObj)
		return
	}

	newPolicy, ok := newObj.(*v1alpha1.AdmissionPolicy)
	if !ok {
		klog.Errorf("unexpected object type: %T", newObj)
		return
	}

	klog.Infof("Updating policy: %s/%s", newPolicy.Namespace, newPolicy.Name)

	// Remove old policy and add new one
	pc.enforcer.RemovePolicy(fmt.Sprintf("%s/%s", oldPolicy.Namespace, oldPolicy.Name))
	
	modelText := newPolicy.Spec.Model
	if newPolicy.Spec.Template != "" {
		if tmpl, ok := casbin.GetTemplate(newPolicy.Spec.Template); ok {
			modelText = tmpl.Model
		}
	}

	if err := pc.enforcer.AddPolicy(
		fmt.Sprintf("%s/%s", newPolicy.Namespace, newPolicy.Name),
		modelText,
		newPolicy.Spec.Rules,
	); err != nil {
		klog.Errorf("failed to update policy: %v", err)
	}
}

func (pc *PolicyController) onDelete(obj interface{}) {
	policy, ok := obj.(*v1alpha1.AdmissionPolicy)
	if !ok {
		klog.Errorf("unexpected object type: %T", obj)
		return
	}

	klog.Infof("Deleting policy: %s/%s", policy.Namespace, policy.Name)
	pc.enforcer.RemovePolicy(fmt.Sprintf("%s/%s", policy.Namespace, policy.Name))
}
