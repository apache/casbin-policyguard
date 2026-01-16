.PHONY: help build test fmt vet docker-build deploy clean

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: fmt vet ## Build manager binary
	go build -o bin/manager cmd/manager/main.go

test: ## Run tests
	go test ./... -v -coverprofile=coverage.out

fmt: ## Run go fmt
	go fmt ./...

vet: ## Run go vet
	go vet ./...

docker-build: ## Build docker image
	docker build -t policywall:latest .

install-crd: ## Install CRDs into the cluster
	kubectl apply -f config/crd/admissionpolicy-crd.yaml

uninstall-crd: ## Uninstall CRDs from the cluster
	kubectl delete -f config/crd/admissionpolicy-crd.yaml

deploy: ## Deploy controller to the cluster
	kubectl apply -f config/webhook/deployment.yaml
	kubectl apply -f config/webhook/webhook-config.yaml

undeploy: ## Undeploy controller from the cluster
	kubectl delete -f config/webhook/webhook-config.yaml
	kubectl delete -f config/webhook/deployment.yaml

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

tidy: ## Run go mod tidy
	go mod tidy
