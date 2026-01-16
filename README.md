# PolicyWall

PolicyWall is a Kubernetes admission webhook that integrates with [Casbin](https://casbin.org/) to provide both validating and mutating admission control for Kubernetes resources. It enables automatic resource compliance through policy-driven mutations and validations.

## Features

- **Mutating Webhook Support**: Automatically modify resources before they are persisted to enforce compliance
  - Sidecar container injection
  - Resource limits enforcement
  - Label/annotation additions
  - Custom JSON Patch operations

- **Validating Webhook Support**: Validate resources against security and compliance policies
  - Security policy enforcement
  - Resource quota validation
  - Custom validation rules

- **Casbin Integration**: Leverage Casbin's powerful policy engine for RBAC/ABAC-based access control

- **CRD-Based Policy Management**: Define policies as Kubernetes custom resources for easy management

- **Conditional Mutations**: Apply mutations based on resource properties and metadata

## Architecture

```
┌─────────────────┐
│   API Server    │
└────────┬────────┘
         │
         │ 1. Admission Request
         ▼
┌─────────────────┐
│  PolicyWall     │
│  Webhook Server │
├─────────────────┤
│ Mutating Hook   │◄──┐
├─────────────────┤   │
│ Validating Hook │   │ 2. Load Policies
├─────────────────┤   │
│ Casbin Enforcer │───┘
└─────────────────┘
         │
         │ 3. Return Response
         │    (Patches/Allow/Deny)
         ▼
```

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured to access your cluster
- cert-manager (for webhook certificates)

### Installation

1. Install the CRD:

```bash
kubectl apply -f config/crd/policy.casbin.org_policies.yaml
```

2. Deploy the webhook:

```bash
kubectl apply -f config/webhook/
```

3. Apply sample policies:

```bash
kubectl apply -f config/samples/
```

## Usage

### Define a Mutation Policy

Create a policy to inject a sidecar container:

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: Policy
metadata:
  name: sidecar-injection-policy
spec:
  subjects:
  - "*"
  resources:
  - resources:
    - pods
    namespaces:
    - default
  mutationRules:
  - name: inject-sidecar-container
    operation: add
    path: /spec/containers/-
    value: |
      {
        "name": "sidecar-proxy",
        "image": "envoyproxy/envoy:v1.27-latest",
        "ports": [{"containerPort": 8080}]
      }
    conditions:
    - field: metadata.labels.inject-sidecar
      operator: equals
      value: "true"
```

### Define a Validation Policy

Create a policy to enforce security constraints:

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: Policy
metadata:
  name: security-validation-policy
spec:
  subjects:
  - "*"
  resources:
  - resources:
    - pods
    namespaces:
    - "*"
  validationRules:
  - name: deny-privileged-containers
    action: deny
    message: "Privileged containers are not allowed"
    conditions:
    - field: spec.containers.0.securityContext.privileged
      operator: equals
      value: "true"
```

### Enable PolicyWall for a Namespace

Label the namespace to enable webhook processing:

```bash
kubectl label namespace default policywall.casbin.org/enabled=true
```

### Test the Webhooks

Create a pod with the sidecar injection label:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  labels:
    inject-sidecar: "true"
spec:
  containers:
  - name: app
    image: nginx:latest
```

The mutating webhook will automatically inject the sidecar container.

## Policy Configuration

### Mutation Rules

Mutation rules define how resources should be modified using JSON Patch operations:

- **operation**: The JSON Patch operation (`add`, `remove`, `replace`)
- **path**: The JSON path to the field to modify
- **value**: The value to set (for `add` and `replace` operations)
- **conditions**: Optional conditions that must be met for the mutation to apply

### Validation Rules

Validation rules define constraints that resources must satisfy:

- **action**: The action to take (`allow` or `deny`)
- **message**: The message to return when validation fails
- **conditions**: Conditions that trigger the validation rule

### Conditions

Conditions support various operators:

- `equals`: Field equals a specific value
- `notEquals`: Field does not equal a specific value
- `in`: Field value is in a list of values
- `notIn`: Field value is not in a list of values
- `exists`: Field exists in the resource
- `notExists`: Field does not exist in the resource

### Resource Selectors

Resource selectors define which Kubernetes resources a policy applies to:

- **apiGroups**: List of API groups (e.g., `[""]`, `["apps"]`)
- **apiVersions**: List of API versions (e.g., `["v1"]`)
- **resources**: List of resource types (e.g., `["pods"]`, `["deployments"]`)
- **namespaces**: List of namespaces (e.g., `["default"]`, `["*"]`)

## Examples

See the `config/samples/` directory for complete examples:

- **sidecar_injection_policy.yaml**: Inject a sidecar container into pods
- **resource_limits_policy.yaml**: Automatically set resource limits and requests
- **security_validation_policy.yaml**: Enforce security best practices

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Run Locally

```bash
make run
```

### Build Docker Image

```bash
make docker-build IMG=your-registry/policywall:tag
```

## Casbin Integration

PolicyWall uses Casbin for policy enforcement. The default model is RBAC with resource-based control:

```ini
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
```

You can customize the Casbin model by setting the `casbinModel` field in the Policy spec.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [Casbin](https://casbin.org/) - Authorization library
- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) - Kubernetes controller framework