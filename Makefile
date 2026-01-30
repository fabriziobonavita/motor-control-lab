SHELL := /bin/bash

.PHONY: help fmt lint test build clean sim-step ci release

.DEFAULT_GOAL := help

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

# Binary name and output directory
BINARY := mcl
CMD_PATH := ./cmd/mcl
BIN_DIR := bin
DIST_DIR := dist
OUT_DIR := artifacts

help: ## Show this help message
	@printf "$(COLOR_BOLD)Motor Control Lab - Available targets:$(COLOR_RESET)\n"
	@printf "\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'
	@printf "\n"

fmt: ## Format Go code
	@printf "$(COLOR_BLUE)Formatting Go code...$(COLOR_RESET)\n"
	@gofmt -s -w .
	@printf "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)\n"

lint: ## Run linters (golangci-lint if available, else go vet)
	@printf "$(COLOR_BLUE)Running linters...$(COLOR_RESET)\n"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		printf "$(COLOR_YELLOW)Note: golangci-lint not found, using go vet instead$(COLOR_RESET)\n"; \
		go vet ./...; \
	fi
	@printf "$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)\n"

test: ## Run tests
	@printf "$(COLOR_BLUE)Running tests...$(COLOR_RESET)\n"
	@go test ./...
	@printf "$(COLOR_GREEN)✓ Tests complete$(COLOR_RESET)\n"

build: ## Build the binary
	@printf "$(COLOR_BLUE)Building $(BINARY)...$(COLOR_RESET)\n"
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR)/$(BINARY) $(CMD_PATH)
	@printf "$(COLOR_GREEN)✓ Built $(BIN_DIR)/$(BINARY)$(COLOR_RESET)\n"

clean: ## Clean build artifacts
	@printf "$(COLOR_BLUE)Cleaning...$(COLOR_RESET)\n"
	@rm -rf $(BIN_DIR) $(DIST_DIR) $(OUT_DIR)
	@go clean -cache -testcache
	@printf "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)\n"

sim-step: build ## Run a step response simulation with default parameters
	@printf "$(COLOR_BLUE)Running step response simulation...$(COLOR_RESET)\n"
	@mkdir -p $(OUT_DIR)
	@$(BIN_DIR)/$(BINARY) sim step \
		--target 1000 \
		--duration 10 \
		--dt 0.001 \
		--kp 0.02 \
		--ki 0.05 \
		--kd 0.0 \
		--deadzone 0.0 \
		--out $(OUT_DIR)
	@printf "$(COLOR_GREEN)✓ Simulation complete. Artifacts in $(OUT_DIR)/$(COLOR_RESET)\n"

ci: fmt test build ## Run CI checks (format, test, build)
	@printf "$(COLOR_GREEN)✓ CI checks passed$(COLOR_RESET)\n"

release: ## Prepare release (checks git state, builds binaries, generates checksums)
	@printf "$(COLOR_BLUE)Preparing release...$(COLOR_RESET)\n"
	@if [ -n "$$(git status --porcelain)" ]; then \
		printf "$(COLOR_YELLOW)Error: Working directory is not clean$(COLOR_RESET)\n"; \
		git status --short; \
		exit 1; \
	fi
	@tag=$$(git describe --tags --exact-match HEAD 2>/dev/null || echo ""); \
	if [ -z "$$tag" ]; then \
		printf "$(COLOR_YELLOW)No git tag found for current commit.$(COLOR_RESET)\n"; \
		printf "To create a release:\n"; \
		printf "  1. git tag vX.Y.Z\n"; \
		printf "  2. git push origin vX.Y.Z\n"; \
		printf "  3. GitHub Actions will build and release automatically\n"; \
		exit 1; \
	fi
	@printf "$(COLOR_GREEN)✓ Release tag: $$tag$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)Building release binaries...$(COLOR_RESET)\n"
	@mkdir -p $(DIST_DIR)
	@version=$${tag#v}; \
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY)_$${tag}_linux_amd64 $(CMD_PATH) && \
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY)_$${tag}_darwin_amd64 $(CMD_PATH) && \
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY)_$${tag}_darwin_arm64 $(CMD_PATH) && \
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY)_$${tag}_windows_amd64.exe $(CMD_PATH) && \
	cd $(DIST_DIR) && sha256sum $(BINARY)_$${tag}_* > SHA256SUMS && \
	printf "$(COLOR_GREEN)✓ Release binaries built in $(DIST_DIR)/$(COLOR_RESET)\n"
