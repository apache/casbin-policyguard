# PolicyWall Integration Examples

This document provides real-world examples of using PolicyWall for various use cases.

## Example 1: Automatic Sidecar Injection

Automatically inject Envoy proxy sidecars into pods with specific labels.

### Policy Definition

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: Policy
metadata:
  name: envoy-sidecar-injection
spec:
  subjects:
  - "*"
  resources:
  - resources:
    - pods
    namespaces:
    - production
    - staging
  mutationRules:
  - name: inject-envoy-sidecar
    operation: add
    path: /spec/containers/-
    value: |
      {
        "name": "envoy-sidecar",
        "image": "envoyproxy/envoy:v1.27-latest",
        "ports": [
          {"containerPort": 8080, "name": "proxy"},
          {"containerPort": 8001, "name": "admin"}
        ],
        "env": [
          {"name": "ENVOY_UID", "value": "0"}
        ],
        "resources": {
          "requests": {"cpu": "100m", "memory": "128Mi"},
          "limits": {"cpu": "200m", "memory": "256Mi"}
        }
      }
    conditions:
    - field: metadata.labels.service-mesh
      operator: equals
      value: "enabled"
  - name: add-sidecar-injected-label
    operation: add
    path: /metadata/labels/sidecar-injected
    value: "true"
    conditions:
    - field: metadata.labels.service-mesh
      operator: equals
      value: "enabled"
```

### Test Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-with-sidecar
  namespace: production
  labels:
    app: myapp
    service-mesh: enabled
spec:
  containers:
  - name: app
    image: myapp:v1.0
    ports:
    - containerPort: 8080
```

## Example 2: Enforce Resource Limits and Quotas

Automatically add resource limits to pods that don't define them.

### Policy Definition

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: Policy
metadata:
  name: enforce-resource-limits
spec:
  subjects:
  - "*"
  resources:
  - resources:
    - pods
    namespaces:
    - production
  mutationRules:
  - name: set-cpu-limits
    operation: add
    path: /spec/containers/0/resources/limits/cpu
    value: "1000m"
    conditions:
    - field: spec.containers.0.resources.limits.cpu
      operator: notExists
  - name: set-memory-limits
    operation: add
    path: /spec/containers/0/resources/limits/memory
    value: "1Gi"
    conditions:
    - field: spec.containers.0.resources.limits.memory
      operator: notExists
  - name: set-cpu-requests
    operation: add
    path: /spec/containers/0/resources/requests/cpu
    value: "100m"
    conditions:
    - field: spec.containers.0.resources.requests.cpu
      operator: notExists
  - name: set-memory-requests
    operation: add
    path: /spec/containers/0/resources/requests/memory
    value: "256Mi"
    conditions:
    - field: spec.containers.0.resources.requests.memory
      operator: notExists
  validationRules:
  - name: enforce-cpu-max
    action: deny
    message: "CPU limit cannot exceed 2 cores"
    conditions:
    - field: spec.containers.0.resources.limits.cpu
      operator: notIn
      values:
      - "100m"
      - "200m"
      - "500m"
      - "1000m"
      - "2000m"
```

## Example 3: Security Policy Enforcement

Prevent insecure pod configurations from being deployed.

### Policy Definition

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: Policy
metadata:
  name: security-hardening
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
    message: "Privileged containers are not allowed for security reasons"
    conditions:
    - field: spec.containers.0.securityContext.privileged
      operator: equals
      value: "true"
  - name: deny-host-network
    action: deny
    message: "Host network access is not allowed"
    conditions:
    - field: spec.hostNetwork
      operator: equals
      value: "true"
  - name: deny-host-pid
    action: deny
    message: "Host PID namespace access is not allowed"
    conditions:
    - field: spec.hostPID
      operator: equals
      value: "true"
  - name: deny-host-ipc
    action: deny
    message: "Host IPC namespace access is not allowed"
    conditions:
    - field: spec.hostIPC
      operator: equals
      value: "true"
  - name: require-security-context
    action: deny
    message: "Pods must define a security context with runAsNonRoot=true"
    conditions:
    - field: spec.securityContext.runAsNonRoot
      operator: notEquals
      value: "true"
```

## Example 4: Multi-Tenant Namespace Isolation

Add network policies and labels based on namespace.

### Policy Definition

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: Policy
metadata:
  name: tenant-isolation
spec:
  subjects:
  - "system:serviceaccount:tenant-a:default"
  resources:
  - resources:
    - pods
    namespaces:
    - tenant-a
  mutationRules:
  - name: add-tenant-label
    operation: add
    path: /metadata/labels/tenant
    value: "tenant-a"
  - name: add-network-policy-label
    operation: add
    path: /metadata/labels/network-policy
    value: "isolated"
  validationRules:
  - name: enforce-tenant-namespace
    action: deny
    message: "Pods can only be created in assigned tenant namespace"
    conditions:
    - field: metadata.namespace
      operator: notEquals
      value: "tenant-a"
```

## Example 5: Image Registry Whitelisting

Ensure only approved container images are used.

### Policy Definition

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: Policy
metadata:
  name: approved-registries
spec:
  subjects:
  - "*"
  resources:
  - resources:
    - pods
    namespaces:
    - production
  validationRules:
  - name: require-approved-registry
    action: deny
    message: "Only images from approved registries are allowed: gcr.io, quay.io, or internal registry"
    conditions:
    - field: spec.containers.0.image
      operator: notIn
      values:
      - "gcr.io/*"
      - "quay.io/*"
      - "registry.internal.company.com/*"
```

## Example 6: Automatic Label and Annotation Management

Add standard labels and annotations to all pods.

### Policy Definition

```yaml
apiVersion: policy.casbin.org/v1alpha1
kind: Policy
metadata:
  name: standard-labels
spec:
  subjects:
  - "*"
  resources:
  - resources:
    - pods
    namespaces:
    - "*"
  mutationRules:
  - name: add-managed-by-label
    operation: add
    path: /metadata/labels/app.kubernetes.io~1managed-by
    value: "policywall"
  - name: add-version-annotation
    operation: add
    path: /metadata/annotations/policywall.casbin.org~1version
    value: "v1"
    conditions:
    - field: metadata.annotations.policywall.casbin.org/version
      operator: notExists
  - name: add-deployment-timestamp
    operation: add
    path: /metadata/annotations/policywall.casbin.org~1injected-at
    value: "{{ .Now }}"
```

## Testing Policies

### Test Command

```bash
kubectl run test-pod \
  --image=nginx:latest \
  --labels=service-mesh=enabled \
  --dry-run=server
```

### Verify Mutations

```bash
kubectl get pod test-pod -o yaml
```

### Check Webhook Logs

```bash
kubectl logs -l app=policywall-webhook -n policywall-system --tail=100
```

## Best Practices

1. **Start with Validation**: Begin with validation rules before adding mutations
2. **Test in Non-Production**: Always test policies in development/staging first
3. **Use Namespaces**: Scope policies to specific namespaces when possible
4. **Monitor Logs**: Keep an eye on webhook logs for policy violations
5. **Version Policies**: Use meaningful names and track policy changes
6. **Document Policies**: Add comments explaining why policies exist
7. **Gradual Rollout**: Enable policies incrementally across namespaces
8. **Backup Policies**: Store policy definitions in version control

## Troubleshooting

### Policy Not Applied

1. Check namespace label: `kubectl get ns <namespace> --show-labels`
2. Verify policy exists: `kubectl get policies`
3. Check webhook logs for errors
4. Verify webhook configuration: `kubectl get mutatingwebhookconfigurations`

### Unexpected Mutations

1. Review policy conditions carefully
2. Check the order of mutations (first match wins)
3. Test with `--dry-run=server` to preview changes
4. Review webhook logs for applied mutations

### Permission Denied

1. Verify RBAC permissions for webhook service account
2. Check cluster role and role bindings
3. Ensure webhook has access to Policy CRD
