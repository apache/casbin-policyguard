# PolicyWall

PolicyWall is a Kubernetes admission webhook controller that enforces access control policies using [Casbin](https://casbin.org/). It provides flexible policy enforcement with support for audit mode (dry-run) to enable safe policy testing in production environments.

## Features

- **Casbin-based Policy Enforcement**: Leverage Casbin's powerful RBAC, ABAC, and other access control models
- **Audit Mode (Dry-Run)**: Test policies in production without disrupting operations
  - Violations are logged and returned as warnings
  - Requests are allowed even when they violate policies
  - Perfect for validating policies before strict enforcement
- **Custom Resource Definition**: Define policies using Kubernetes-native CRDs
- **Flexible Matching Rules**: Apply policies selectively based on API groups, resources, and operations
- **Real-time Policy Updates**: Policies are dynamically loaded without restart
- **Multiple Policy Support**: Run multiple policies with different configurations simultaneously

## Architecture

PolicyWall consists of two main components:

1. **Webhook Server**: Validates admission requests against loaded policies
2. **Controller**: Watches AdmissionPolicy CRDs and updates the webhook server

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.19+)
- kubectl configured to access your cluster
- cert-manager (for webhook certificates) or manually generated certificates

### Installation

1. Install the CRD:
```bash
kubectl apply -f config/crd/admissionpolicy-crd.yaml
```

2. Deploy the controller:
```bash
kubectl apply -f config/webhook/deployment.yaml
```

3. Configure the webhook (update caBundle with your CA certificate):
```bash
kubectl apply -f config/webhook/webhook-config.yaml
```

### Creating a Policy

#### Dry-Run Mode (Audit Only)

Create a policy in dry-run mode to test without enforcement:

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: AdmissionPolicy
metadata:
  name: test-policy
spec:
  # Enable dry-run mode for safe testing
  dryRun: true
  
  model: |
    [request_definition]
    r = sub, obj, act
    
    [policy_definition]
    p = sub, obj, act
    
    [role_definition]
    g = _, _
    
    [policy_effect]
    e = some(where (p.eft == allow))
    
    [matchers]
    m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
  
  policy: |
    p, role:admin, production/*, DELETE
    g, admin@example.com, role:admin
  
  matchRules:
  - apiGroups: [""]
    resources: ["pods", "services"]
    operations: ["DELETE"]
```

When a user tries to delete a pod in the `production` namespace:
- If they have permission: Request is allowed
- If they don't have permission in dry-run mode:
  - Request is **still allowed**
  - A warning is added to the response
  - Violation is logged by the webhook server

#### Enforcement Mode

Switch to enforcement mode by setting `dryRun: false`:

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: AdmissionPolicy
metadata:
  name: strict-policy
spec:
  # Strict enforcement
  dryRun: false
  
  model: |
    [request_definition]
    r = sub, obj, act
    
    [policy_definition]
    p = sub, obj, act
    
    [policy_effect]
    e = some(where (p.eft == allow))
    
    [matchers]
    m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
  
  policy: |
    p, admin@example.com, production/critical-app, DELETE
```

In enforcement mode:
- Violations **deny the request**
- Non-compliant operations are blocked
- Status is returned with the denial reason

### Checking Policy Status

```bash
kubectl get admissionpolicies
```

Output:
```
NAME           DRYRUN   READY   VIOLATIONS   AGE
test-policy    true     true    42           5m
strict-policy  false    true    0            2m
```

### Workflow: From Dry-Run to Enforcement

1. **Create policy in dry-run mode**:
   ```bash
   kubectl apply -f config/samples/dryrun-policy.yaml
   ```

2. **Monitor violations** in logs and warnings:
   ```bash
   kubectl logs -n policywall-system deployment/policywall-controller
   ```

3. **Adjust policy** based on observed violations

4. **Switch to enforcement** when ready:
   ```bash
   kubectl patch admissionpolicy test-policy --type=merge -p '{"spec":{"dryRun":false}}'
   ```

## Configuration

### AdmissionPolicy CRD Spec

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `dryRun` | bool | No | Enable audit mode (default: false) |
| `model` | string | Yes | Casbin model definition |
| `policy` | string | Yes | Casbin policy rules |
| `matchRules` | []MatchRule | No | Rules to match admission requests |

### MatchRule Spec

| Field | Type | Description |
|-------|------|-------------|
| `apiGroups` | []string | API groups to match (use "*" for all) |
| `apiVersions` | []string | API versions to match |
| `resources` | []string | Resource types to match |
| `operations` | []string | Operations to match (CREATE, UPDATE, DELETE, CONNECT) |

## Development

### Building

```bash
make build
```

### Running Tests

```bash
make test
```

### Building Docker Image

```bash
make docker-build
```

## Examples

See the [examples](config/samples/) directory for more policy examples:
- [dryrun-policy.yaml](config/samples/dryrun-policy.yaml) - Audit mode example
- [enforce-policy.yaml](config/samples/enforce-policy.yaml) - Strict enforcement example

## Security Considerations

- **Namespace Exclusion**: The webhook excludes namespaces with the label `policy.casbin.org/ignore`
- **Failure Policy**: Set to `Fail` by default for security (can be changed to `Ignore`)
- **Certificate Management**: Use cert-manager or ensure certificates are properly rotated
- **Audit Logs**: Review dry-run violations before enabling enforcement

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.