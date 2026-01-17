# policywall

PolicyWall is a Kubernetes admission controller that uses [Casbin](https://casbin.org/) for policy enforcement. It provides a built-in web dashboard for policy visualization, editing, and testing.

## Features

- **📊 Overview Dashboard**: Real-time metrics showing request counts, violation rates, and active policies
- **✏️ Policy Editor**: Graphical form-based editor for creating Kubernetes AdmissionPolicy YAML resources
- **🎮 Playground**: Safe sandbox environment for testing policies against admission requests using Casbin enforcer

## Installation

```bash
# Build from source
go build -o policywall ./cmd/policywall
```

## Usage

### Start the Web Dashboard

```bash
# Start dashboard on default port 8080
./policywall dashboard

# Start dashboard on custom port
./policywall dashboard --port 9090
```

Then open your browser to `http://localhost:8080` to access the dashboard.

### Dashboard Features

#### Overview Page
- View cluster-wide metrics (total requests, allowed/denied counts, violation rates)
- Monitor active policies
- Quick links to create new policies or test existing ones

#### Policy Editor
- Create AdmissionPolicy resources using an intuitive form
- Automatically generate YAML from form inputs
- Validate policy syntax before deployment
- Copy generated YAML to clipboard

#### Playground
- Test policies safely without affecting your live cluster
- Paste JSON admission requests to evaluate against policies
- Uses the Casbin enforcer for real-time policy evaluation
- Try example requests to understand policy behavior

## Example Policy

The playground comes pre-configured with an RBAC model:

```
# Admin can access all resources
p, admin, /*, *

# Users can GET resources under /api/
p, user, /api/*, GET

# Role assignments
g, alice, admin
g, bob, user
```

Test it with example requests:
```json
{
  "subject": "alice",
  "object": "/api/pods",
  "action": "GET"
}
```

## License

Apache License 2.0