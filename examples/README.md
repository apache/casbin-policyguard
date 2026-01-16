# PolicyWall Examples

This directory contains example AdmissionPolicy configurations demonstrating various use cases.

## Files

- **dryrun-policy.yaml**: Example of audit mode (dry-run) policy for safe testing
- **enforce-policy.yaml**: Example of strict enforcement policy
- **rbac-example.yaml**: Complete RBAC example with multiple roles
- **workflow-example.yaml**: Example workflow from dry-run to enforcement

## Quick Start

1. Apply the CRD:
   ```bash
   kubectl apply -f ../crd/admissionpolicy-crd.yaml
   ```

2. Start with a dry-run policy to test:
   ```bash
   kubectl apply -f dryrun-policy.yaml
   ```

3. Monitor the logs for violations:
   ```bash
   kubectl logs -n policywall-system deployment/policywall-controller -f
   ```

4. When satisfied, switch to enforcement:
   ```bash
   kubectl apply -f enforce-policy.yaml
   ```

## Dry-Run Mode

When `dryRun: true` is set:
- Violations are logged as warnings
- Requests are allowed even if they violate the policy
- Warnings appear in kubectl output
- Perfect for production testing without disruption

## Enforcement Mode

When `dryRun: false` is set:
- Violations deny the request
- Non-compliant operations are blocked
- Status includes the denial reason
