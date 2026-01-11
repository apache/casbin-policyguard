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

func BenchmarkEnforcer_Enforce(b *testing.B) {
	enforcer := NewPolicyEnforcer()

	model := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

	rules := []string{
		"alice, data1, read",
		"bob, data2, write",
		"charlie, data3, read",
	}

	if err := enforcer.AddPolicy("benchmark-policy", model, rules); err != nil {
		b.Fatalf("failed to add policy: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = enforcer.Enforce("benchmark-policy", "alice", "data1", "read")
	}
}

func BenchmarkEnforcer_AddPolicy(b *testing.B) {
	model := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`

	rules := []string{"alice, data1, read"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enforcer := NewPolicyEnforcer()
		_ = enforcer.AddPolicy("test-policy", model, rules)
	}
}
