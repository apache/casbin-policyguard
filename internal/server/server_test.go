package server

import (
	"testing"
)


func TestNewServer(t *testing.T) {
srv := NewServer("8080")
if srv == nil {
t.Fatal("NewServer returned nil")
}
if srv.port != "8080" {
t.Errorf("Expected port 8080, got %s", srv.port)
}
if srv.enforcer == nil {
t.Error("Enforcer should not be nil")
}
}

func TestGetMetrics(t *testing.T) {
srv := NewServer("8080")
metrics := srv.getMetrics()

if metrics == nil {
t.Fatal("getMetrics returned nil")
}

// Check that metrics contain expected keys
expectedKeys := []string{"totalRequests", "allowedRequests", "deniedRequests", "violationRate", "activePolicies"}
for _, key := range expectedKeys {
if _, ok := metrics[key]; !ok {
t.Errorf("Metrics missing expected key: %s", key)
}
}
}
