# Observability & Validation Improvements

This document describes the observability and validation improvements added to PolicyWall based on user feedback.

## Overview

Four key improvements have been implemented:

1. **Prometheus Metrics** - Real-time monitoring of policy violations
2. **Detailed Error Messages** - Specific rule context in denial messages
3. **Status Tracking** - Recent violations visible in CRD status
4. **Model Validation** - Validating webhook prevents invalid Casbin models

## 1. Prometheus Metrics

### Available Metrics

#### `policywall_policy_violations_total`
Tracks total policy violations detected.

**Labels:**
- `policy`: Policy name
- `mode`: `dryrun` or `enforce`
- `result`: `allowed` (dry-run) or `denied` (enforce)
- `namespace`: Resource namespace
- `resource`: Resource kind (Pod, Deployment, etc.)

**Example:**
```promql
# Dry-run violations in the last hour
rate(policywall_policy_violations_total{mode="dryrun"}[1h])

# Denied requests by policy
sum by (policy) (policywall_policy_violations_total{result="denied"})
```

#### `policywall_admission_requests_total`
Tracks all admission requests processed.

**Labels:**
- `operation`: CREATE, UPDATE, DELETE, CONNECT
- `namespace`: Resource namespace
- `resource`: Resource kind

**Example:**
```promql
# Request rate by operation
rate(policywall_admission_requests_total[5m]) by (operation)
```

#### `policywall_policy_evaluation_duration_seconds`
Histogram of policy evaluation time.

**Labels:**
- `policy`: Policy name

**Example:**
```promql
# 95th percentile evaluation time
histogram_quantile(0.95, rate(policywall_policy_evaluation_duration_seconds_bucket[5m]))
```

#### `policywall_active_policies`
Gauge of active policies loaded.

**Labels:**
- `mode`: `dryrun` or `enforce`

**Example:**
```promql
# Number of active policies by mode
policywall_active_policies
```

### Accessing Metrics

Metrics are exposed on the metrics endpoint (default `:8080/metrics`):

```bash
curl http://localhost:8080/metrics | grep policywall
```

## 2. Detailed Error Messages

### Before
```
Policy 'production-policy' violation: DELETE Pod/myapp in namespace production not allowed
```

### After
```
Policy 'production-policy' denied: user 'john@example.com' cannot DELETE Pod 'production/myapp' in namespace 'production'. Required: policy rule matching (subject=john@example.com, object=production/myapp, action=DELETE)
```

### Benefits
- **User identification**: See who attempted the action
- **Resource details**: Full resource path with namespace
- **Rule requirements**: Understand what policy rule would be needed
- **Debugging aid**: Quickly identify misconfigured policies

## 3. Status Tracking

### CRD Status Enhancement

The AdmissionPolicy status now includes:

```yaml
status:
  ready: true
  message: "Policy loaded in dry-run mode (audit only). 15 violations detected."
  violationCount: 15
  lastUpdated: "2026-01-16T17:00:00Z"
  recentViolations:
  - kind: Pod
    namespace: production
    name: critical-app
    operation: DELETE
    timestamp: "2026-01-16T16:59:45Z"
    user: developer@example.com
    message: "Policy 'production-policy' denied: user 'developer@example.com'..."
  - kind: Deployment
    namespace: staging
    name: web-app
    operation: UPDATE
    timestamp: "2026-01-16T16:58:30Z"
    user: ops@example.com
    message: "Policy 'namespace-policy' denied: user 'ops@example.com'..."
```

### Viewing Status

```bash
# Quick view
kubectl get admissionpolicies

NAME                 DRYRUN   READY   VIOLATIONS   AGE
production-policy    true     true    15           2h

# Detailed view
kubectl get admissionpolicy production-policy -o yaml

# Watch violations in real-time
kubectl get admissionpolicy production-policy -o jsonpath='{.status.recentViolations[*].message}' -w
```

### Benefits
- **Direct visibility**: No log parsing required
- **Real-time updates**: Status updates automatically
- **Resource tracking**: See exactly which resources violated
- **User audit trail**: Track who attempted what operations
- **Limited retention**: Only last 10 violations to prevent unbounded growth

## 4. Model Validation

### Validating Webhook

A validating webhook now catches invalid Casbin models at admission time:

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: AdmissionPolicy
metadata:
  name: invalid-policy
spec:
  model: "invalid syntax here"  # Will be rejected
  policy: "p, admin"
```

**Result:**
```
Error from server: admission webhook "validate-admissionpolicy.policy.casbin.org" denied the request: invalid Casbin model: syntax error at line 1
```

### Validation Rules

1. **Model required**: `spec.model` cannot be empty
2. **Model syntax**: Must be valid Casbin model syntax
3. **Empty policy warning**: Warns if `spec.policy` is empty
4. **Match rules warning**: Warns if match rules have no filters
5. **Dry-run recommendation**: Warns when creating policy with `dryRun=false`

### Examples

**Invalid model - rejected:**
```bash
$ kubectl apply -f invalid-policy.yaml
Error: invalid Casbin model: section not found: [request_definition]
```

**Empty policy - warning:**
```bash
$ kubectl apply -f empty-policy.yaml
Warning: spec.policy is empty - this policy will deny all requests
admissionpolicy.policy.casbin.org/test created
```

**New enforcement policy - warning:**
```bash
$ kubectl apply -f enforce-policy.yaml
Warning: Creating policy with dryRun=false. Consider testing with dryRun=true first.
admissionpolicy.policy.casbin.org/prod-policy created
```

### Benefits
- **Early error detection**: Catch issues before deployment
- **Prevent silent failures**: No more policies that fail at runtime
- **Best practices**: Encourages dry-run testing
- **Immediate feedback**: See errors in kubectl output

## Comparison with OPA Gatekeeper

| Feature | PolicyWall (Before) | PolicyWall (Now) | OPA Gatekeeper |
|---------|---------------------|------------------|----------------|
| Prometheus metrics | ❌ | ✅ | ✅ |
| Detailed error messages | ❌ | ✅ | ✅ |
| Violation tracking in status | ❌ | ✅ | ✅ |
| Validating webhook | ❌ | ✅ | ✅ |
| Dry-run mode | ✅ | ✅ | ✅ |
| Casbin integration | ✅ | ✅ | ❌ |
