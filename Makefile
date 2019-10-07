# Go parameters
GOCMD=go
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=rbac-manager
VERSION := "dev"
COMMIT := $(shell git rev-parse HEAD)

all: test
test:
	printf "Linter:\n"
	$(GOCMD) list ./... | xargs -L1 golint | tee golint-report.out
	printf "\n\nTests:\n\n"
	$(GOCMD) test -v -coverprofile coverage.txt -covermode=atomic ./...
	$(GOCMD) vet ./... 2> govet-report.out
	$(GOCMD) tool cover -html=coverage.txt -o cover-report.html
	printf "\nCoverage report available at cover-report.html\n\n"
clean:
	$(GOCLEAN)
	$(GOCMD) fmt ./...
	rm -f $(BINARY_NAME)
build:
	$(GOCMD) build -ldflags "-w -s -X main.version=$(VERSION) -X main.commit=$(COMMIT)" -a -o rbac-manager ./cmd/manager/main.go
build-linux: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOCMD) build -ldflags="-w -s" -a -o rbac-manager ./cmd/manager/main.go 
