# PolicyWall

[![CI](https://github.com/casbin/policywall/workflows/CI/badge.svg)](https://github.com/casbin/policywall/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/casbin/policywall)](https://goreportcard.com/report/github.com/casbin/policywall)
[![Coverage Status](https://codecov.io/gh/casbin/policywall/branch/master/graph/badge.svg)](https://codecov.io/gh/casbin/policywall)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub release](https://img.shields.io/github/release/casbin/policywall.svg)](https://github.com/casbin/policywall/releases)

A Kubernetes-native admission policy controller powered by [Casbin](https://casbin.org). PolicyWall provides a flexible, easy-to-install admission control solution for Kubernetes clusters with support for custom policies, audit mode, and comprehensive metrics.

## Features

- **🚀 Easy Installation**: Simple Helm chart deployment with minimal configuration
- **📋 CRD-Based Policies**: Kubernetes-native policy management using Custom Resource Definitions
- **🔒 Policy Templates**: Pre-built templates for common scenarios:
  - Pod security policies
  - Image tag validation
  - Resource quota enforcement
  - Namespace isolation
- **🔍 Audit Mode**: Dry-run capability to test policies without enforcement
- **📊 Metrics & Monitoring**: Prometheus metrics for operational visibility
- **⚡ High Performance**: Optimized for low-latency admission decisions
- **🛡️ Secure by Default**: TLS-enabled webhook with security best practices

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured to access your cluster
- Helm 3.x

### Installation

1. **Install PolicyWall using Helm:**

```bash
helm repo add policywall https://casbin.github.io/policywall
helm repo update
helm install policywall policywall/policywall -n policywall-system --create-namespace
```

2. **Verify the installation:**

```bash
kubectl get pods -n policywall-system
kubectl get crd admissionpolicies.policywall.casbin.org
```

3. **Apply your first policy:**

```bash
kubectl apply -f examples/pod-security-policy.yaml
```

### Example: Pod Security Policy

Create a policy to prevent privileged pods in most namespaces:

```yaml
apiVersion: policywall.casbin.org/v1alpha1
kind: AdmissionPolicy
metadata:
  name: pod-security-policy
  namespace: default
spec:
  template: pod-security
  rules:
    - "p, default, privileged, deny"
    - "p, kube-system, privileged, allow"
  failurePolicy: Fail
  dryRun: false
  namespaceSelector:
    matchLabels:
      policy: enabled
```

### Example: Image Validation Policy

Enforce approved container registries and prevent `latest` tags:

```yaml
apiVersion: policywall.casbin.org/v1alpha1
kind: AdmissionPolicy
metadata:
  name: image-validation-policy
  namespace: default
spec:
  template: image-validation
  rules:
    - "p, *, latest, deny"
    - "p, *, gcr.io, allow"
    - "p, *, docker.io, allow"
  failurePolicy: Fail
  dryRun: false
```

## Policy Templates

PolicyWall includes built-in templates for common use cases:

| Template | Description |
|----------|-------------|
| `pod-security` | Enforce pod security standards (privileged containers, host namespaces, etc.) |
| `image-validation` | Validate container images (registries, tags, signatures) |
| `resource-quota` | Enforce resource limits and quotas |
| `namespace-isolation` | Control cross-namespace resource access |

List available templates using the CLI:

```bash
policywall templates
```

## CLI Tool

PolicyWall includes a CLI tool for policy management and auditing:

```bash
# Audit existing resources against policies
policywall audit --namespace default

# Audit all namespaces
policywall audit

# List available templates
policywall templates

# Check version
policywall version
```

## Dry-Run Mode

Test policies before enforcing them:

```yaml
spec:
  dryRun: true  # Violations will be logged but not blocked
```

## Monitoring

PolicyWall exposes Prometheus metrics on port 9090:

- `policywall_admission_requests_total` - Total admission requests
- `policywall_policy_evaluation_duration_seconds` - Policy evaluation latency

Access metrics:

```bash
kubectl port-forward -n policywall-system svc/policywall 9090:9090
curl http://localhost:9090/metrics
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/casbin/policywall.git
cd policywall

# Build the controller
go build -o bin/controller ./cmd/controller

# Build the CLI
go build -o bin/policywall ./cmd/cli

# Run tests
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

### Running Locally

```bash
# Run the controller locally (requires kubeconfig)
go run cmd/controller/main.go --kubeconfig ~/.kube/config
```

## Architecture

PolicyWall consists of:

1. **Admission Webhook Server**: Intercepts CREATE/UPDATE operations
2. **Policy Controller**: Watches AdmissionPolicy CRDs and updates Casbin enforcer
3. **Casbin Enforcer**: Evaluates requests against policies
4. **Metrics Collector**: Exposes Prometheus metrics

```
┌─────────────────┐
│   API Server    │
└────────┬────────┘
         │
         │ AdmissionReview
         ▼
┌─────────────────┐      ┌──────────────────┐
│ Webhook Server  │─────▶│ Policy Enforcer  │
└─────────────────┘      └──────────────────┘
         │                        │
         │                        │ Watch
         ▼                        ▼
┌─────────────────┐      ┌──────────────────┐
│    Metrics      │      │ AdmissionPolicy  │
│   Collector     │      │      CRDs        │
└─────────────────┘      └──────────────────┘
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## License

Copyright 2026 The Casbin Authors. All Rights Reserved.

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Community

- Website: https://casbin.org
- GitHub: https://github.com/casbin/policywall
- Forum: https://forum.casbin.com
- Gitter: https://gitter.im/casbin/lobby

## Acknowledgments

Built with:
- [Casbin](https://casbin.org) - Authorization library
- [Kubernetes](https://kubernetes.io) - Container orchestration
- [Prometheus](https://prometheus.io) - Monitoring and alerting