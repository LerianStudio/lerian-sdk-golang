# Lerian Go SDK Makefile
#
# Enterprise-grade build system for the multi-product Lerian SDK.
# Run 'make help' for a summary of all available targets.

# ======================================================================
# Variables
# ======================================================================

SERVICE_NAME := Lerian Go SDK
BIN_DIR      := ./bin
ARTIFACTS_DIR := ./artifacts
DOCS_DIR     := ./docs/godoc
VERSION      := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Ensure output directories exist
$(shell mkdir -p $(ARTIFACTS_DIR))

# Go toolchain
GO       := go
GOFMT    := gofmt
GOLINT   := golangci-lint
GOMOD    := $(GO) mod
GOBUILD  := $(GO) build
GOTEST   := $(GO) test
GOTOOL   := $(GO) tool
GOCLEAN  := $(GO) clean

# Project
PROJECT_ROOT := $(shell pwd)
PROJECT_NAME := lerian-go-sdk
MODULE       := $(shell $(GO) list -m)

# Environment
ENV_FILE         := $(PROJECT_ROOT)/.env
ENV_EXAMPLE_FILE := $(PROJECT_ROOT)/.env.example

# Load environment variables if .env exists
ifneq (,$(wildcard .env))
    include .env
endif

# Helper for section headers
define print_header
	@echo ""
	@echo "==== $(1) ===="
	@echo ""
endef

# ======================================================================
# Core Commands
# ======================================================================

.PHONY: help
help: ## Display this help message
	@echo ""
	@echo "$(SERVICE_NAME) Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  make %-22s %s\n", $$1, $$2}'
	@echo ""

.PHONY: set-env
set-env: ## Create .env file from .env.example
	$(call print_header,Setting up environment)
	@if [ ! -f "$(ENV_FILE)" ] && [ -f "$(ENV_EXAMPLE_FILE)" ]; then \
		echo "No .env file found. Creating from .env.example..."; \
		cp $(ENV_EXAMPLE_FILE) $(ENV_FILE); \
		echo "[ok] Created .env file from .env.example"; \
	elif [ ! -f "$(ENV_FILE)" ] && [ ! -f "$(ENV_EXAMPLE_FILE)" ]; then \
		echo "[error] Neither .env nor .env.example files found"; \
		exit 1; \
	elif [ -f "$(ENV_FILE)" ]; then \
		read -t 10 -p ".env file already exists. Overwrite with .env.example? [Y/n] (auto-yes in 10s) " answer || answer="Y"; \
		answer=$${answer:-Y}; \
		if [[ $$answer =~ ^[Yy] ]]; then \
			cp $(ENV_EXAMPLE_FILE) $(ENV_FILE); \
			echo "[ok] Overwrote .env file with .env.example"; \
		else \
			echo "[skipped] Kept existing .env file"; \
		fi; \
	fi

# ======================================================================
# Build
# ======================================================================

.PHONY: build
build: ## Build all packages in the module
	$(call print_header,Building packages)
	@$(GOBUILD) ./...
	@echo "[ok] Build completed successfully"

# ======================================================================
# Test Commands
# ======================================================================

.PHONY: test test-fast coverage

test: ## Run the full test suite with race detection
	$(call print_header,Running tests)
	@$(GOTEST) -race -v ./...

test-fast: ## Run tests in short mode (skip long-running tests)
	$(call print_header,Running fast tests)
	@$(GOTEST) -short -race ./...

coverage: ## Generate HTML coverage report in artifacts/
	$(call print_header,Generating test coverage)
	@$(GOTEST) -race -coverprofile=$(ARTIFACTS_DIR)/coverage.out \
		$$($(GO) list ./... | grep -v -E '(examples|mocks)')
	@$(GOTOOL) cover -html=$(ARTIFACTS_DIR)/coverage.out -o $(ARTIFACTS_DIR)/coverage.html
	@echo "Coverage report: $(ARTIFACTS_DIR)/coverage.html"
	@echo "[ok] Coverage report generated successfully"

# ======================================================================
# Code Quality
# ======================================================================

.PHONY: lint fmt tidy gosec verify-sdk

lint: ## Run golangci-lint with the full linter suite
	$(call print_header,Running linters)
	@if ! command -v $(GOLINT) > /dev/null; then \
		echo "Installing golangci-lint..."; \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@$(GOLINT) run ./...
	@echo "[ok] Linting completed successfully"

fmt: ## Format all Go source files
	$(call print_header,Formatting code)
	@$(GOFMT) -s -w .
	@echo "[ok] Formatting completed successfully"

tidy: ## Tidy module dependencies
	$(call print_header,Tidying dependencies)
	@$(GOMOD) tidy
	@echo "[ok] Dependencies tidied successfully"

gosec: ## Run gosec security scanner
	$(call print_header,Running security checks)
	@if ! command -v gosec > /dev/null; then \
		echo "Installing gosec..."; \
		$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@gosec -quiet ./...
	@echo "[ok] Security checks completed successfully"

verify-sdk: ## Quick build + vet check to verify SDK compiles cleanly
	$(call print_header,Verifying SDK)
	@$(GOBUILD) ./...
	@$(GO) vet ./...
	@echo "[ok] SDK verified successfully"

# ======================================================================
# Example Commands
# ======================================================================

.PHONY: example-midaz example-multi

example-midaz: ## Run the Midaz workflow example
	$(call print_header,Running Midaz Workflow Example)
	@cd examples/midaz-workflow && $(GO) run .

example-multi: ## Run the multi-product example
	$(call print_header,Running Multi-Product Example)
	@cd examples/multi-product && $(GO) run .

# ======================================================================
# Documentation
# ======================================================================

.PHONY: godoc

godoc: ## Start a godoc server at http://localhost:6060
	$(call print_header,Starting godoc server)
	@echo "Starting godoc server at http://localhost:6060/pkg/$(MODULE)/"
	@if ! command -v godoc > /dev/null; then \
		echo "Installing godoc..."; \
		$(GO) install golang.org/x/tools/cmd/godoc@latest; \
	fi
	@godoc -http=:6060

# ======================================================================
# Clean
# ======================================================================

.PHONY: clean

clean: ## Remove build artifacts and coverage reports
	$(call print_header,Cleaning build artifacts)
	@$(GOCLEAN)
	@rm -rf $(BIN_DIR)/ $(ARTIFACTS_DIR)/
	@echo "[ok] Artifacts cleaned successfully"
