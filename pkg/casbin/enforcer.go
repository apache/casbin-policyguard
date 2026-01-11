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

package casbin

import (
	"fmt"
	"strings"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

// PolicyEnforcer manages Casbin enforcers for admission policies
type PolicyEnforcer struct {
	enforcers map[string]*casbin.Enforcer
	mu        sync.RWMutex
}

// NewPolicyEnforcer creates a new PolicyEnforcer
func NewPolicyEnforcer() *PolicyEnforcer {
	return &PolicyEnforcer{
		enforcers: make(map[string]*casbin.Enforcer),
	}
}

// AddPolicy adds a new policy with the given name and Casbin model
func (pe *PolicyEnforcer) AddPolicy(name string, modelText string, rules []string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	e, err := casbin.NewEnforcer(m)
	if err != nil {
		return fmt.Errorf("failed to create enforcer: %w", err)
	}

	for _, rule := range rules {
		parts := parseRule(rule)
		params := make([]interface{}, len(parts))
		for i, p := range parts {
			params[i] = p
		}
		if _, err := e.AddPolicy(params...); err != nil {
			return fmt.Errorf("failed to add rule: %w", err)
		}
	}

	pe.enforcers[name] = e
	return nil
}

// RemovePolicy removes a policy by name
func (pe *PolicyEnforcer) RemovePolicy(name string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	delete(pe.enforcers, name)
}

// Enforce checks if the request is allowed by the named policy
func (pe *PolicyEnforcer) Enforce(name string, params ...interface{}) (bool, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	e, ok := pe.enforcers[name]
	if !ok {
		return false, fmt.Errorf("policy %s not found", name)
	}

	return e.Enforce(params...)
}

// parseRule parses a policy rule string into components
func parseRule(rule string) []string {
	// Simple comma-separated parsing
	var parts []string
	for _, part := range strings.Split(rule, ",") {
		parts = append(parts, strings.TrimSpace(part))
	}
	return parts
}
