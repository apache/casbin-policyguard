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

// PolicyTemplate defines a reusable policy template
type PolicyTemplate struct {
	Name        string
	Description string
	Model       string
	DefaultRules []string
}

// Templates contains built-in policy templates
var Templates = map[string]PolicyTemplate{
	"pod-security": {
		Name:        "pod-security",
		Description: "Enforce pod security standards",
		Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
		DefaultRules: []string{
			"p, *, privileged, deny",
			"p, system:serviceaccount:kube-system, privileged, allow",
		},
	},
	"image-validation": {
		Name:        "image-validation",
		Description: "Validate container image tags and registries",
		Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
		DefaultRules: []string{
			"p, *, latest, deny",
			"p, *, trusted-registry, allow",
		},
	},
	"resource-quota": {
		Name:        "resource-quota",
		Description: "Enforce resource quotas and limits",
		Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
		DefaultRules: []string{
			"p, *, no-limits, deny",
			"p, *, with-limits, allow",
		},
	},
	"namespace-isolation": {
		Name:        "namespace-isolation",
		Description: "Enforce namespace isolation policies",
		Model: `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`,
		DefaultRules: []string{
			"p, *, cross-namespace, deny",
			"p, admin, cross-namespace, allow",
		},
	},
}

// GetTemplate returns a template by name
func GetTemplate(name string) (PolicyTemplate, bool) {
	tmpl, ok := Templates[name]
	return tmpl, ok
}
