# Workflow: From Dry-Run to Enforcement

This example demonstrates a complete workflow for implementing a new policy safely using audit mode.

## Step 1: Deploy in Dry-Run Mode

```bash
kubectl apply -f - <<EOF
apiVersion: policy.casbin.org/v1alpha1
kind: AdmissionPolicy
metadata:
  name: production-protection
spec:
  # Enable dry-run mode for testing
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
    # Only SRE team can delete in production
    p, role:sre, production/*, DELETE
    g, sre-team@example.com, role:sre
  
  matchRules:
  - apiGroups: ["", "apps"]
    resources: ["*"]
    operations: ["DELETE"]
EOF
```

## Step 2: Monitor Violations

Watch the logs to see what would be blocked. Since the webhook runs as part of the controller:

```bash
kubectl logs -n policywall-system deployment/policywall-controller -f | grep "DRY-RUN"
```

You'll see output like:
```
W0116 01:19:54.677559 [DRY-RUN] Policy 'production-protection' violation: DELETE Pod/myapp in namespace production not allowed
```

Note: The DRY-RUN violations are logged by the webhook handler within the controller pod.

## Step 3: Identify False Positives

Review the violations and adjust your policy if needed:

```bash
kubectl edit admissionpolicy production-protection
```

Update the policy rules to handle legitimate cases.

## Step 4: Enable Enforcement

Once satisfied with the policy, disable dry-run mode:

```bash
kubectl patch admissionpolicy production-protection --type=merge -p '{"spec":{"dryRun":false}}'
```

## Step 5: Verify Enforcement

Try to delete a resource without proper permissions:

```bash
kubectl delete pod myapp -n production
```

You should see:
```
Error from server: admission webhook "validate.policy.casbin.org" denied the request: Request denied by policies: [Policy 'production-protection' violation: DELETE Pod/myapp in namespace production not allowed]
```

## Step 6: Rollback if Needed

If you encounter issues, quickly rollback to dry-run mode:

```bash
kubectl patch admissionpolicy production-protection --type=merge -p '{"spec":{"dryRun":true}}'
```

## Best Practices

1. **Always start with dry-run**: Never enable enforcement without testing first
2. **Monitor for at least 24-48 hours**: Capture different usage patterns
3. **Review all warnings**: Don't assume they're all expected
4. **Document exceptions**: Keep track of why certain violations are acceptable
5. **Gradual rollout**: Start with less critical namespaces
6. **Have a rollback plan**: Know how to quickly disable enforcement if needed
