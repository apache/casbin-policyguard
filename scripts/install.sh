#!/bin/bash
# Copyright 2026 The Casbin Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

NAMESPACE="${NAMESPACE:-policywall-system}"
RELEASE_NAME="${RELEASE_NAME:-policywall}"

echo "Installing PolicyWall..."
echo "Namespace: $NAMESPACE"
echo "Release: $RELEASE_NAME"

# Create namespace
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Install CRD
echo "Installing CRD..."
kubectl apply -f config/crd/admissionpolicy.yaml

# Generate TLS certificates for webhook
echo "Generating TLS certificates..."
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# Generate CA
openssl genrsa -out ca.key 2048
openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=PolicyWall CA"

# Generate server certificate
openssl genrsa -out tls.key 2048
openssl req -new -key tls.key -out tls.csr -subj "/CN=policywall.${NAMESPACE}.svc"
openssl x509 -req -days 365 -in tls.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out tls.crt

# Create secret
kubectl create secret tls "${RELEASE_NAME}-certs" \
  --cert=tls.crt \
  --key=tls.key \
  -n "$NAMESPACE" \
  --dry-run=client -o yaml | kubectl apply -f -

# Get CA bundle
CA_BUNDLE=$(cat ca.crt | base64 | tr -d '\n')

cd -
rm -rf "$TEMP_DIR"

# Install using Helm
if command -v helm &> /dev/null; then
  echo "Installing with Helm..."
  helm upgrade --install "$RELEASE_NAME" ./deploy/helm/policywall \
    --namespace "$NAMESPACE" \
    --set webhook.caBundle="$CA_BUNDLE"
else
  echo "Helm not found. Please install Helm or apply resources manually."
  exit 1
fi

echo "PolicyWall installed successfully!"
echo ""
echo "Next steps:"
echo "1. Verify installation: kubectl get pods -n $NAMESPACE"
echo "2. Create a policy: kubectl apply -f examples/pod-security-policy.yaml"
echo "3. Check the quickstart guide: docs/QUICKSTART.md"
