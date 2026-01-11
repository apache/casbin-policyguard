# Contributing to PolicyWall

Thank you for your interest in contributing to PolicyWall! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in the [Issues](https://github.com/casbin/policywall/issues)
2. If not, create a new issue with:
   - Clear description of the problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (K8s version, Go version, etc.)

### Suggesting Features

1. Check existing feature requests in Issues
2. Create a new issue with the `enhancement` label
3. Describe the feature and its use case
4. Discuss the implementation approach

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with appropriate tests
4. Ensure all tests pass (`make test`)
5. Run linters (`make lint`)
6. Commit with conventional commits (e.g., `feat:`, `fix:`, `docs:`)
7. Push to your fork and create a pull request

## Development Setup

### Prerequisites

- Go 1.23.0 or later
- Docker (for building images)
- Kubernetes cluster (for testing)
- kubectl
- Helm 3.x

### Building

```bash
# Clone the repository
git clone https://github.com/casbin/policywall.git
cd policywall

# Install dependencies
make deps

# Build binaries
make build

# Run tests
make test

# Run linters
make lint
```

### Running Locally

```bash
# Run the controller (requires kubeconfig)
go run cmd/controller/main.go --kubeconfig ~/.kube/config
```

## Coding Standards

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Add comments for exported functions and types
- Write tests for new functionality
- Maintain test coverage above 80%

## Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
```
feat(webhook): add support for DELETE operations
fix(enforcer): handle nil policy gracefully
docs(readme): update installation instructions
```

## Testing

- Write unit tests for all new code
- Use table-driven tests where appropriate
- Mock external dependencies
- Ensure tests are deterministic

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
