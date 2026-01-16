# PolicyWall Deployment Guide

This guide shows how to deploy PolicyWall to a Kubernetes cluster.

## Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured to access your cluster
- cert-manager installed (for automatic certificate generation)

## Install cert-manager

If not already installed:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

## Generate Webhook Certificates

Create a certificate issuer and certificate:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: policywall-selfsigned-issuer
  namespace: policywall-system
spec:
  selfSigned: {}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: policywall-webhook-cert
  namespace: policywall-system
spec:
  secretName: policywall-webhook-cert
  dnsNames:
  - policywall-webhook-service.policywall-system.svc
  - policywall-webhook-service.policywall-system.svc.cluster.local
  issuerRef:
    name: policywall-selfsigned-issuer
```

Save this as `config/deployment/certificate.yaml` and apply:

```bash
kubectl apply -f config/deployment/certificate.yaml
```

## Deploy PolicyWall

1. Install the CRD:

```bash
kubectl apply -f config/crd/policy.casbin.org_policies.yaml
```

2. Deploy the webhook server:

```bash
kubectl apply -f config/deployment/deployment.yaml
```

3. Wait for the webhook to be ready:

```bash
kubectl wait --for=condition=ready pod -l app=policywall-webhook -n policywall-system --timeout=60s
```

4. Configure the webhook (requires CA bundle):

First, get the CA bundle from the certificate:

```bash
export CA_BUNDLE=$(kubectl get secret policywall-webhook-cert -n policywall-system -o jsonpath='{.data.ca\.crt}')
```

Then update the webhook configurations to include the CA bundle:

```bash
# For mutating webhook
cat config/webhook/mutating_webhook_configuration.yaml | \
  sed "s/caBundle: .*/caBundle: ${CA_BUNDLE}/" | \
  kubectl apply -f -

# For validating webhook
cat config/webhook/validating_webhook_configuration.yaml | \
  sed "s/caBundle: .*/caBundle: ${CA_BUNDLE}/" | \
  kubectl apply -f -
```

Alternatively, if using cert-manager's CA injector annotation, update the webhook configurations to include:

```yaml
metadata:
  annotations:
    cert-manager.io/inject-ca-from: policywall-system/policywall-webhook-cert
```

## Enable PolicyWall for Namespaces

Label namespaces where you want PolicyWall to be active:

```bash
kubectl label namespace default policywall.casbin.org/enabled=true
```

## Apply Sample Policies

```bash
kubectl apply -f config/samples/
```

## Test the Deployment

Create a test pod to verify mutations and validations:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
  labels:
    inject-sidecar: "true"
spec:
  containers:
  - name: app
    image: nginx:latest
```

Check that the sidecar was injected:

```bash
kubectl get pod test-pod -o jsonpath='{.spec.containers[*].name}'
# Should show: app sidecar-proxy
```

## Troubleshooting

### Check webhook logs

```bash
kubectl logs -l app=policywall-webhook -n policywall-system
```

### Verify webhook is registered

```bash
kubectl get mutatingwebhookconfigurations
kubectl get validatingwebhookconfigurations
```

### Check certificate status

```bash
kubectl get certificate -n policywall-system
kubectl describe certificate policywall-webhook-cert -n policywall-system
```

### Test webhook connectivity

```bash
kubectl run test-pod --image=nginx --dry-run=server
```

This will trigger the webhook without creating the pod. Check the webhook logs for any errors.

## Uninstall

```bash
kubectl delete -f config/webhook/
kubectl delete -f config/deployment/
kubectl delete -f config/crd/
kubectl delete namespace policywall-system
```
