SHELL := /bin/bash

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=rbac-manager
COMMIT := $(shell git rev-parse HEAD)
VERSION := "dev"

SYSTEM                := $(shell uname -s | tr A-Z a-z)_$(shell uname -m | sed "s/x86_64/amd64/")
GOLANGCI_LINT_VERSION := 1.19.1

tools/golangci-lint:
	@mkdir -p tools
	@curl -sSLf \
		https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/golangci-lint-$(GOLANGCI_LINT_VERSION)-$(shell echo $(SYSTEM) | tr '_' '-').tar.gz \
		| tar xzOf - golangci-lint-$(GOLANGCI_LINT_VERSION)-$(shell echo $(SYSTEM) | tr '_' '-')/golangci-lint > tools/golangci-lint && chmod +x tools/golangci-lint

all: lint test

.PHONY: lint
lint: tools/golangci-lint
	./tools/golangci-lint run ./...

.PHONY: test
test:
	printf "\n\nTests:\n\n"
	$(GOCMD) test -v -coverprofile coverage.txt -covermode=atomic ./...
	$(GOCMD) vet ./... 2> govet-report.out
	$(GOCMD) tool cover -html=coverage.txt -o cover-report.html
	printf "\nCoverage report available at cover-report.html\n\n"

.PHONY: tidy
tidy:
	$(GOCMD) mod tidy

.PHONY: clean
clean:
	$(GOCLEAN)
	$(GOCMD) fmt ./...
	rm -f $(BINARY_NAME)
	rm -rf tools
