# Image URL to use all building/pushing image targets
IMG ?= policywall:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: fmt vet ## Build webhook binary.
	go build -o bin/policywall main.go

.PHONY: run
run: fmt vet ## Run webhook from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: ## Build docker image with the webhook.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the webhook.
	docker push ${IMG}

##@ Deployment

.PHONY: install
install: ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd/

.PHONY: uninstall
uninstall: ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/crd/

.PHONY: deploy
deploy: ## Deploy webhook to the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/webhook/

.PHONY: undeploy
undeploy: ## Undeploy webhook from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/webhook/

##@ Sample

.PHONY: samples
samples: ## Apply sample policies
	kubectl apply -f config/samples/
