# Makefile for EgenSkriven
# 
# Usage:
#   make dev          - Start development server with hot reload
#   make build        - Build production binary with embedded UI
#   make test         - Run all tests
#   make test-coverage - Run tests with coverage report
#   make clean        - Remove build artifacts
#   make build-ui     - Build React UI for production
#   make dev-ui       - Start React dev server
#   make dev-all      - Start both React and Go dev servers

.PHONY: dev build run clean test test-coverage tidy help build-ui dev-ui test-ui clean-ui dev-all

# Default target: show help
help:
	@echo "Available commands:"
	@echo "  make dev           - Start development server with hot reload"
	@echo "  make build         - Build production binary with embedded UI"
	@echo "  make run           - Build and run the server"
	@echo "  make test          - Run all tests (Go + UI)"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Remove build artifacts and data"
	@echo "  make tidy          - Tidy Go module dependencies"
	@echo ""
	@echo "UI commands:"
	@echo "  make build-ui      - Build React UI for production"
	@echo "  make dev-ui        - Start React dev server (port 5173)"
	@echo "  make test-ui       - Run UI tests"
	@echo "  make clean-ui      - Remove UI build artifacts"
	@echo "  make dev-all       - Start both React and Go dev servers"

# Development: run with hot reload using Air
# Requires: go install github.com/air-verse/air@latest
dev:
	@echo "Starting development server with hot reload..."
	@echo "Install Air if missing: go install github.com/air-verse/air@latest"
	air

# Build production binary with embedded UI
# CGO_ENABLED=0 ensures pure Go build (no C dependencies)
# This is important for cross-platform compatibility
build: build-ui
	@echo "Building Go binary with embedded UI..."
	CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven
	@echo "Built: ./egenskriven ($$(du -h egenskriven | cut -f1))"

# Build and run the application
run: build
	@echo "Starting server..."
	./egenskriven serve

# Clean build artifacts and data
clean: clean-ui
	@echo "Cleaning build artifacts..."
	rm -rf egenskriven
	rm -rf pb_data/
	rm -rf .air/
	rm -rf coverage.out coverage.html
	@echo "Clean complete"

# Run all tests with verbose output
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests with coverage report
# Generates both console output and HTML report
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	@echo ""
	@echo "Coverage summary:"
	go tool cover -func=coverage.out
	@echo ""
	go tool cover -html=coverage.out -o coverage.html
	@echo "HTML report generated: coverage.html"

# Tidy dependencies (remove unused, add missing)
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "Done"

# =============================================================================
# UI Development
# =============================================================================

# Build the React UI for production
build-ui:
	@echo "Building React UI..."
	cd ui && npm ci && npm run build
	@echo "UI built: ui/dist/"

# Start React development server (port 5173)
# Use this alongside 'make dev' for hot reload on both frontend and backend
dev-ui:
	@echo "Starting React dev server..."
	@echo "Note: Run 'make dev' in another terminal for the Go backend"
	cd ui && npm run dev

# Run UI tests
test-ui:
	@echo "Running UI tests..."
	cd ui && npm test

# Clean UI build artifacts
clean-ui:
	@echo "Cleaning UI artifacts..."
	rm -rf ui/dist ui/node_modules

# =============================================================================
# Combined Development
# =============================================================================

# Development: run both React and Go with hot reload
# This runs both servers in parallel using Make's -j flag
dev-all:
	@echo "Starting development servers..."
	@echo "  React: http://localhost:5173 (with proxy to :8090)"
	@echo "  Go:    http://localhost:8090"
	@$(MAKE) -j2 dev-ui dev
