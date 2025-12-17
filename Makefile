.PHONY: all build test lint clean install help

# Build variables
BINARY_NAME := journal2day1
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go variables
GOBIN := $(shell go env GOPATH)/bin
GOLANGCI_LINT := $(GOBIN)/golangci-lint

# Default target
all: lint test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/journal2day1

# Run tests
test:
	@echo "Running tests..."
	go test -race -cover ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -race -cover -v ./...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install golangci-lint if not present
$(GOLANGCI_LINT):
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
lint: $(GOLANGCI_LINT)
	@echo "Running linter..."
	$(GOLANGCI_LINT) run ./...

# Run linter with auto-fix
lint-fix: $(GOLANGCI_LINT)
	@echo "Running linter with auto-fix..."
	$(GOLANGCI_LINT) run --fix ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(GOBIN)..."
	cp bin/$(BINARY_NAME) $(GOBIN)/

# Run the converter (example usage)
run: build
	./bin/$(BINARY_NAME) convert -i ~/AppleJournalEntries -o ~/dayone-import.zip

# Show help
help:
	@echo "Available targets:"
	@echo "  all           - Run lint, test, and build (default)"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run golangci-lint"
	@echo "  lint-fix      - Run golangci-lint with auto-fix"
	@echo "  fmt           - Format code"
	@echo "  tidy          - Tidy dependencies"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  run           - Build and run with example input"
	@echo "  help          - Show this help"
