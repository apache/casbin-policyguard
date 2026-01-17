package policy

import (
	"encoding/json"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

type Enforcer struct {
	casbin *casbin.Enforcer
}

func NewEnforcer() *Enforcer {
	// Create a basic RBAC model with wildcard support
	m, _ := model.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && (p.act == "*" || r.act == p.act)
`)

	e, _ := casbin.NewEnforcer(m)
	
	// Add some default policies for demo
	e.AddPolicy("admin", "/*", "*")
	e.AddPolicy("user", "/api/*", "GET")
	e.AddGroupingPolicy("alice", "admin")
	e.AddGroupingPolicy("bob", "user")

	return &Enforcer{
		casbin: e,
	}
}

func (e *Enforcer) EvaluateRequest(requestJSON string, policyStr string) (bool, error) {
	// Parse the request JSON
	var req map[string]interface{}
	if err := json.Unmarshal([]byte(requestJSON), &req); err != nil {
		return false, fmt.Errorf("invalid request JSON: %v", err)
	}

	// Extract subject, object, action from the request
	subject, _ := req["subject"].(string)
	object, _ := req["object"].(string)
	action, _ := req["action"].(string)

	if subject == "" || object == "" || action == "" {
		return false, fmt.Errorf("request must contain subject, object, and action fields")
	}

	// For playground, we use the existing enforcer
	// In a real implementation, we would create a temporary enforcer with the provided policy
	allowed, err := e.casbin.Enforce(subject, object, action)
	if err != nil {
		return false, fmt.Errorf("policy enforcement error: %v", err)
	}

	return allowed, nil
}

func (e *Enforcer) GetPolicies() [][]string {
	policies, _ := e.casbin.GetPolicy()
	return policies
}
