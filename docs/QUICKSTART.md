# PolicyWall Quickstart Guide

This guide will help you get PolicyWall up and running in 5 minutes.

## Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured
- Helm 3.x installed

## Installation

### Step 1: Install PolicyWall

```bash
# Add the PolicyWall Helm repository
helm repo add policywall https://casbin.github.io/policywall
helm repo update

# Install PolicyWall
helm install policywall policywall/policywall \
  --namespace policywall-system \
  --create-namespace
```

### Step 2: Verify Installation

```bash
# Check that PolicyWall is running
kubectl get pods -n policywall-system

# Verify the CRD is installed
kubectl get crd admissionpolicies.policywall.casbin.org
```

### Step 3: Create Your First Policy

Create a file named `first-policy.yaml`:

```yaml
apiVersion: policywall.casbin.org/v1alpha1
kind: AdmissionPolicy
metadata:
  name: deny-privileged-pods
  namespace: default
spec:
  template: pod-security
  rules:
    - "p, default, privileged, deny"
    - "p, kube-system, privileged, allow"
  failurePolicy: Fail
  dryRun: true  # Start in dry-run mode
  namespaceSelector:
    matchLabels:
      policy: enabled
```

Apply the policy:

```bash
kubectl apply -f first-policy.yaml
```

### Step 4: Test the Policy

Enable the policy on a namespace:

```bash
kubectl label namespace default policy=enabled
```

Try creating a privileged pod (this should be logged but not blocked in dry-run mode):

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-privileged
  namespace: default
spec:
  containers:
  - name: nginx
    image: nginx:1.21
    securityContext:
      privileged: true
```

```bash
kubectl apply -f test-pod.yaml
```

Check the PolicyWall logs:

```bash
kubectl logs -n policywall-system -l app.kubernetes.io/name=policywall
```

### Step 5: Enable Enforcement

Once you've verified the policy works as expected, disable dry-run mode:

```bash
kubectl patch admissionpolicy deny-privileged-pods \
  -p '{"spec":{"dryRun":false}}' \
  --type=merge
```

Now try creating the privileged pod again - it should be rejected!

## Next Steps

- Explore other [policy templates](../README.md#policy-templates)
- Set up [monitoring](../README.md#monitoring) with Prometheus
- Use the [CLI tool](../README.md#cli-tool) for auditing
- Read the [full documentation](../README.md)

## Cleanup

```bash
# Uninstall PolicyWall
helm uninstall policywall -n policywall-system

# Remove the namespace
kubectl delete namespace policywall-system

# Delete policies
kubectl delete admissionpolicies --all
```

## Troubleshooting

### Policy not enforcing

1. Check that the namespace has the correct label
2. Verify the policy exists: `kubectl get admissionpolicy`
3. Check PolicyWall logs for errors

### Webhook not responding

1. Verify the service is running: `kubectl get svc -n policywall-system`
2. Check pod health: `kubectl get pods -n policywall-system`
3. Review webhook configuration: `kubectl get validatingwebhookconfiguration`

### Need Help?

- [GitHub Issues](https://github.com/casbin/policywall/issues)
- [Casbin Forum](https://forum.casbin.com)
- [Gitter Chat](https://gitter.im/casbin/lobby)
