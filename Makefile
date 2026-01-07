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

.PHONY: dev build run clean test test-coverage tidy help build-ui dev-ui test-ui clean-ui dev-all \
        release clean-dist release-darwin release-linux release-windows checksums

# Version info (can be overridden: make build VERSION=1.0.0)
VERSION ?= dev
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Linker flags for embedding version info
LDFLAGS := -X github.com/ramtinJ95/EgenSkriven/internal/commands.Version=$(VERSION)
LDFLAGS += -X github.com/ramtinJ95/EgenSkriven/internal/commands.BuildDate=$(BUILD_DATE)
LDFLAGS += -X github.com/ramtinJ95/EgenSkriven/internal/commands.GitCommit=$(GIT_COMMIT)

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
	@echo ""
	@echo "Release commands:"
	@echo "  make release       - Build binaries for all platforms"
	@echo "  make clean-dist    - Clean dist directory"

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
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o egenskriven ./cmd/egenskriven
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

# =============================================================================
# Release Targets
# =============================================================================

# Build binaries for all platforms
# Usage: make release
# Creates binaries in dist/ directory
release: clean-dist build-ui release-darwin release-linux release-windows checksums
	@echo "Release builds complete. Binaries in dist/"
	@ls -lh dist/

# Clean dist directory before release build
clean-dist:
	@echo "Cleaning dist directory..."
	rm -rf dist/
	mkdir -p dist/

# macOS builds (Apple Silicon and Intel)
release-darwin:
	@echo "Building for macOS (arm64)..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-darwin-arm64 \
		./cmd/egenskriven
	@echo "Building for macOS (amd64)..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-darwin-amd64 \
		./cmd/egenskriven

# Linux builds (amd64 and arm64)
release-linux:
	@echo "Building for Linux (amd64)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-linux-amd64 \
		./cmd/egenskriven
	@echo "Building for Linux (arm64)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-linux-arm64 \
		./cmd/egenskriven

# Windows build
release-windows:
	@echo "Building for Windows (amd64)..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags "$(LDFLAGS)" \
		-o dist/egenskriven-windows-amd64.exe \
		./cmd/egenskriven

# Generate checksums for all binaries (cross-platform compatible)
checksums:
	@echo "Generating checksums..."
	cd dist && shasum -a 256 * > checksums.txt
	@cat dist/checksums.txt
