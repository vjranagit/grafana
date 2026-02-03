.PHONY: build test clean run-oncall run-flow install lint

# Build variables
BINARY_NAME=grafana-ops
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/grafana-ops

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -f *.db *.sqlite *.sqlite3

# Run oncall server locally
run-oncall:
	@echo "Starting oncall server..."
	go run ./cmd/grafana-ops oncall --debug

# Run flow agent locally
run-flow:
	@echo "Starting flow agent..."
	go run ./cmd/grafana-ops flow --debug

# Install binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/grafana-ops

# Run linters
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download

# Generate mocks (if needed)
generate:
	@echo "Running go generate..."
	go generate ./...

# Run all checks (test, lint, fmt)
check: fmt lint test

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/grafana-ops
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/grafana-ops
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/grafana-ops
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/grafana-ops

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t grafana-ops:$(VERSION) .

# Help target
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Remove build artifacts"
	@echo "  run-oncall    - Run oncall server locally"
	@echo "  run-flow      - Run flow agent locally"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  tidy          - Tidy dependencies"
	@echo "  deps          - Download dependencies"
	@echo "  check         - Run fmt, lint, and test"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  docker-build  - Build Docker image"

.DEFAULT_GOAL := build
