.PHONY: help test test-unit test-integration test-coverage test-verbose test-clean test-recipe test-all clean build run docker-up docker-down

# Default target
help:
	@echo "Ultra-Bis Makefile Commands:"
	@echo "  make test              - Run all tests (unit + integration)"
	@echo "  make test-all          - Run all tests from all packages (CI/CD)"
	@echo "  make test-unit         - Run unit tests only (fast, no database)"
	@echo "  make test-integration  - Run integration tests only (with test containers)"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-verbose      - Run tests with verbose output"
	@echo "  make test-clean        - Run tests with clean output (no verbose logs)"
	@echo "  make test-recipe       - Run recipe package tests only"
	@echo "  make clean             - Clean build artifacts and test cache"
	@echo "  make build             - Build the application binary"
	@echo "  make run               - Run the application locally"
	@echo "  make docker-up         - Start application with Docker Compose"
	@echo "  make docker-down       - Stop Docker Compose services"

# Run all tests
test:
	@echo "Running all tests..."
	go test -v -timeout 3m ./...

# Run only unit tests (fast tests without database)
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

# Run only integration tests (with test containers)
test-integration:
	@echo "Running integration tests..."
	go test -v -run Integration -timeout 3m ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -timeout 3m ./...

# Clean build artifacts and test cache
clean:
	@echo "Cleaning..."
	go clean
	go clean -testcache
	rm -f coverage.out coverage.html
	rm -rf bin/

# Build the application
build:
	@echo "Building application..."
	go build -o bin/api cmd/api/main.go
	@echo "Binary created: bin/api"

# Run the application locally
run:
	@echo "Running application..."
	go run cmd/api/main.go

# Start Docker Compose services
docker-up:
	@echo "Starting Docker Compose services..."
	docker-compose up -d --build

# Stop Docker Compose services
docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Quick test (auth and food only for fast feedback)
test-quick:
	@echo "Running quick tests (auth + food)..."
	go test -v -timeout 1m ./internal/auth ./internal/food

# Run tests with clean output (no verbose Docker logs)
test-clean:
	@echo "Running tests with clean output..."
	@go test -timeout 3m ./... 2>&1 | grep -E '(^(PASS|FAIL|ok|--- PASS|--- FAIL)|Test)' || true
	@echo "Done. Use 'make test' for verbose output."

# Run recipe package tests only
test-recipe:
	@echo "Running recipe tests..."
	@go test -timeout 2m ./internal/recipe/tests
	@echo "Done."

# Run all tests from all packages (used by CI/CD)
test-all:
	@echo "=========================================="
	@echo "Running ALL tests from all packages..."
	@echo "=========================================="
	@echo ""
	@failed=0; \
	for pkg in internal/auth internal/barcode internal/database internal/diary internal/food internal/goal internal/httputil internal/metrics internal/middleware internal/recipe internal/user; do \
		echo "📦 Testing $$pkg..."; \
		if go test -timeout 2m ./$$pkg/tests 2>&1; then \
			echo "✅ $$pkg tests PASSED"; \
		else \
			echo "❌ $$pkg tests FAILED"; \
			failed=$$((failed + 1)); \
		fi; \
		echo ""; \
	done; \
	echo "=========================================="; \
	if [ $$failed -eq 0 ]; then \
		echo "✅ All package tests PASSED!"; \
	else \
		echo "❌ $$failed package(s) had failures"; \
		exit 1; \
	fi; \
	echo "=========================================="
