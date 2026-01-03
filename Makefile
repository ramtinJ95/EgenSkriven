# Makefile for EgenSkriven
# 
# Usage:
#   make dev          - Start development server with hot reload
#   make build        - Build production binary
#   make test         - Run all tests
#   make test-coverage - Run tests with coverage report
#   make clean        - Remove build artifacts

.PHONY: dev build run clean test test-coverage tidy help

# Default target: show help
help:
	@echo "Available commands:"
	@echo "  make dev           - Start development server with hot reload"
	@echo "  make build         - Build production binary"
	@echo "  make run           - Build and run the server"
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Remove build artifacts and data"
	@echo "  make tidy          - Tidy Go module dependencies"

# Development: run with hot reload using Air
# Requires: go install github.com/air-verse/air@latest
dev:
	@echo "Starting development server with hot reload..."
	@echo "Install Air if missing: go install github.com/air-verse/air@latest"
	air

# Build production binary
# CGO_ENABLED=0 ensures pure Go build (no C dependencies)
# This is important for cross-platform compatibility
build:
	@echo "Building production binary..."
	CGO_ENABLED=0 go build -o egenskriven ./cmd/egenskriven
	@echo "Built: ./egenskriven ($$(du -h egenskriven | cut -f1))"

# Build and run the application
run: build
	@echo "Starting server..."
	./egenskriven serve

# Clean build artifacts and data
clean:
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
