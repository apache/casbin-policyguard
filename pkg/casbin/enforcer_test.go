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
	"testing"
)

func TestNewPolicyEnforcer(t *testing.T) {
	enforcer := NewPolicyEnforcer()
	if enforcer == nil {
		t.Fatal("expected non-nil enforcer")
	}
	if enforcer.enforcers == nil {
		t.Fatal("expected non-nil enforcers map")
	}
}

func TestAddPolicy(t *testing.T) {
	enforcer := NewPolicyEnforcer()

	model := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

	rules := []string{"alice, data1, read"}

	err := enforcer.AddPolicy("test-policy", model, rules)
	if err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	// Verify policy was added
	if _, ok := enforcer.enforcers["test-policy"]; !ok {
		t.Fatal("policy was not added to enforcers map")
	}
}

func TestRemovePolicy(t *testing.T) {
	enforcer := NewPolicyEnforcer()

	model := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

	rules := []string{"alice, data1, read"}

	err := enforcer.AddPolicy("test-policy", model, rules)
	if err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	enforcer.RemovePolicy("test-policy")

	// Verify policy was removed
	if _, ok := enforcer.enforcers["test-policy"]; ok {
		t.Fatal("policy was not removed from enforcers map")
	}
}

func TestEnforce(t *testing.T) {
	enforcer := NewPolicyEnforcer()

	model := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

	rules := []string{"alice, data1, read"}

	err := enforcer.AddPolicy("test-policy", model, rules)
	if err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	// Test enforcement
	allowed, err := enforcer.Enforce("test-policy", "alice", "data1", "read")
	if err != nil {
		t.Fatalf("failed to enforce: %v", err)
	}
	if !allowed {
		t.Fatal("expected policy to allow request")
	}

	// Test non-existent policy
	_, err = enforcer.Enforce("non-existent", "alice", "data1", "read")
	if err == nil {
		t.Fatal("expected error for non-existent policy")
	}
}
