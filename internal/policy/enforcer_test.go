package policy

import (
	"testing"
)


func TestNewEnforcer(t *testing.T) {
e := NewEnforcer()
if e == nil {
t.Fatal("NewEnforcer returned nil")
}
if e.casbin == nil {
t.Fatal("Casbin enforcer is nil")
}
}

func TestEvaluateRequest(t *testing.T) {
e := NewEnforcer()

tests := []struct {
name     string
request  string
expected bool
}{
{
name:     "admin can access all",
request:  `{"subject": "alice", "object": "/api/pods", "action": "GET"}`,
expected: true,
},
{
name:     "user can GET /api/*",
request:  `{"subject": "bob", "object": "/api/pods", "action": "GET"}`,
expected: true,
},
{
name:     "user cannot DELETE",
request:  `{"subject": "bob", "object": "/api/secrets", "action": "DELETE"}`,
expected: false,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
allowed, err := e.EvaluateRequest(tt.request, "")
if err != nil {
t.Fatalf("EvaluateRequest failed: %v", err)
}
if allowed != tt.expected {
t.Errorf("Expected %v, got %v", tt.expected, allowed)
}
})
}
}

func TestEvaluateRequestInvalidJSON(t *testing.T) {
e := NewEnforcer()
_, err := e.EvaluateRequest("invalid json", "")
if err == nil {
t.Error("Expected error for invalid JSON, got nil")
}
}

func TestEvaluateRequestMissingFields(t *testing.T) {
e := NewEnforcer()
_, err := e.EvaluateRequest(`{"subject": "alice"}`, "")
if err == nil {
t.Error("Expected error for missing fields, got nil")
}
}

func TestGetPolicies(t *testing.T) {
e := NewEnforcer()
policies := e.GetPolicies()
if len(policies) == 0 {
t.Error("Expected some policies, got none")
}
}
