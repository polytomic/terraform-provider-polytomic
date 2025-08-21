# Makefile for round-trip tests

.PHONY: test test-round-trip test-verbose test-connection test-model test-sync test-example clean help

# Default target
help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Environment variables for tests
export TF_ACC = 1
export TF_LOG_PROVIDER = DEBUG

# Run all round-trip tests
test: ## Run all round-trip tests
	go test -v -timeout 30m ./roundtrip

# Build tests without running
build: ## Build tests without running them
	go build ./roundtrip

# Clean up test artifacts
clean: ## Clean up temporary files and artifacts
	go clean -testcache
	rm -rf /tmp/roundtrip-*

# Run tests in parallel (use with caution - may hit API rate limits)
test-parallel: ## Run tests in parallel (careful with API limits)
	go test -v -timeout 45m -parallel 2 ./roundtrip

# Run specific test by name
test-run: ## Run specific test (usage: make test-run TEST=TestName)
ifndef TEST
	@echo "Usage: make test-run TEST=TestAccRoundTrip_Connection"
	@exit 1
endif
	go test -v -timeout 15m -run "$(TEST)" ./roundtrip

# Lint test code
lint: ## Run linter on test code
	golangci-lint run ./roundtrip/...

# Format test code
fmt: ## Format test code
	go fmt ./roundtrip/...

# Check for missing environment variables
check-env: ## Check for required environment variables
	@echo "Checking required environment variables..."
	@test -n "$(POLYTOMIC_API_KEY)" || test -n "$(POLYTOMIC_DEPLOYMENT_KEY)" || (echo "POLYTOMIC_API_KEY or POLYTOMIC_DEPLOYMENT_KEY must be set" && exit 1)
	@test -n "$(POLYTOMIC_DEPLOYMENT_URL)" || (echo "POLYTOMIC_DEPLOYMENT_URL must be set" && exit 1)
	@echo "Environment variables OK"

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	go test -v -timeout 30m -coverprofile=coverage.out ./roundtrip
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Integration with main provider tests
test-all: check-env ## Run all tests including provider tests
	cd .. && make testacc
	$(MAKE) test

# Quick test - just run the basic connection test
test-quick: ## Quick test - basic connection only
	go test -v -timeout 5m -run "TestAccRoundTrip_Connection" ./roundtrip

# Debug test - run with maximum logging
test-debug: ## Run tests with debug logging
	TEST_KEEP_WORKSPACES=true TF_LOG=TRACE TF_LOG_PROVIDER=TRACE go test -v -timeout 30m ./roundtrip