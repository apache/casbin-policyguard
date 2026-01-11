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

.PHONY: all build test clean lint fmt vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Binary names
CONTROLLER_BINARY=bin/controller
CLI_BINARY=bin/policywall

all: test build

build: build-controller build-cli

build-controller:
	$(GOBUILD) -o $(CONTROLLER_BINARY) -v ./cmd/controller

build-cli:
	$(GOBUILD) -o $(CLI_BINARY) -v ./cmd/cli

test:
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

bench:
	$(GOTEST) -bench=. -benchmem ./...

clean:
	rm -f $(CONTROLLER_BINARY) $(CLI_BINARY)
	rm -f coverage.txt

lint:
	golangci-lint run

fmt:
	$(GOFMT) ./...

vet:
	$(GOVET) ./...

deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

help:
	@echo "Available targets:"
	@echo "  all            - Run tests and build"
	@echo "  build          - Build all binaries"
	@echo "  build-controller - Build controller binary"
	@echo "  build-cli      - Build CLI binary"
	@echo "  test           - Run tests with coverage"
	@echo "  bench          - Run benchmarks"
	@echo "  clean          - Remove built binaries"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  deps           - Download dependencies"
