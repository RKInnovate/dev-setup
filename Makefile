# File: Makefile
# Purpose: Build automation for devsetup Go binary
# Problem: Need consistent commands for building, testing, and installing
# Role: Provides make targets for common development tasks
# Usage: Run `make` or `make build` to compile, `make install` to install globally
# Design choices: Uses Go build with version injection; supports multiple architectures
# Assumptions: Go 1.21+ installed; running on macOS

.PHONY: all build install clean test lint help

# Version info (injected at build time)
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION := $(GIT_TAG)+$(GIT_COMMIT)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Binary name
BINARY_NAME := devsetup

# Default target
all: build

## build: Build the devsetup binary for current architecture
build:
	@echo "Building $(BINARY_NAME) (version: $(VERSION))..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/devsetup
	@echo "✅ Built: ./$(BINARY_NAME)"

## build-all: Build binaries for all supported architectures
build-all:
	@echo "Building for all architectures..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/devsetup
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/devsetup
	@echo "✅ Built:"
	@echo "  - $(BINARY_NAME)-darwin-amd64 (Intel Mac)"
	@echo "  - $(BINARY_NAME)-darwin-arm64 (Apple Silicon)"

## install: Build and install to ~/.local/bin
install: build
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@mkdir -p ~/.local/bin
	@cp $(BINARY_NAME) ~/.local/bin/$(BINARY_NAME)
	@chmod +x ~/.local/bin/$(BINARY_NAME)
	@echo "✅ Installed: ~/.local/bin/$(BINARY_NAME)"
	@echo ""
	@echo "Ensure ~/.local/bin is in your PATH:"
	@echo '  export PATH="$$HOME/.local/bin:$$PATH"'

## test: Run all tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "✅ Tests passed"

## lint: Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./... && \
		echo "✅ Linting passed"; \
	else \
		echo "⚠️  golangci-lint not installed. Install with:"; \
		echo "    brew install golangci-lint"; \
	fi

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*
	rm -f coverage.out
	@echo "✅ Cleaned"

## run: Build and run with install command
run: build
	@echo "Running $(BINARY_NAME) install..."
	./$(BINARY_NAME) install

## dry-run: Build and run in dry-run mode
dry-run: build
	@echo "Running $(BINARY_NAME) install --dry-run..."
	./$(BINARY_NAME) install --dry-run

## help: Show this help message
help:
	@echo "devsetup Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
