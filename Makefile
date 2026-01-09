# scmd Makefile

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/scmd/scmd/pkg/version.Version=$(VERSION) \
	-X github.com/scmd/scmd/pkg/version.Commit=$(COMMIT) \
	-X github.com/scmd/scmd/pkg/version.Date=$(DATE)

.PHONY: all build test lint clean install dev fmt vet coverage deps help \
	release release-snapshot release-dry-run tag completions docker

# Default target
all: lint test build

# Build the binary
build:
	@echo "Building scmd..."
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/scmd ./cmd/scmd

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-darwin-arm64 ./cmd/scmd
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-darwin-amd64 ./cmd/scmd
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-linux-amd64 ./cmd/scmd
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-linux-arm64 ./cmd/scmd
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-windows-amd64.exe ./cmd/scmd

# Run tests
test:
	@echo "Running tests..."
	go test -race -coverprofile=coverage.out ./...

# Run tests with short flag (for CI)
test-short:
	@echo "Running short tests (excluding e2e)..."
	go test -short -race $(shell go list ./... | grep -v '/tests/e2e')

# Run tests with verbose output
test-v:
	@echo "Running tests (verbose)..."
	go test -race -v ./...

# Generate coverage report
coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, running go vet only"; \
		go vet ./...; \
	fi

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

# Run in development mode
dev:
	@echo "Running scmd..."
	go run ./cmd/scmd

# Install to /usr/local/bin
install: build
	@echo "Installing scmd..."
	cp bin/scmd /usr/local/bin/
	@echo "Installed to /usr/local/bin/scmd"

# Install to GOPATH/bin
install-go: build
	@echo "Installing scmd to GOPATH..."
	go install ./cmd/scmd

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/ dist/ coverage.out coverage.html completions/ npm/bin/ npm/tmp/

# Tidy dependencies
deps:
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify

# Generate (placeholder for future code generation)
generate:
	@echo "Running go generate..."
	go generate ./...

# Check for outdated dependencies
outdated:
	@echo "Checking for outdated dependencies..."
	go list -u -m all

# Security audit
audit:
	@echo "Running security audit..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Generate shell completions
completions:
	@echo "Generating shell completions..."
	@mkdir -p completions
	@go run ./cmd/scmd completion bash > completions/scmd.bash
	@go run ./cmd/scmd completion zsh > completions/scmd.zsh
	@go run ./cmd/scmd completion fish > completions/scmd.fish
	@echo "Completions generated in completions/"

# Create a new git tag
tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tag created. Push with: git push origin $(VERSION)"

# Release with GoReleaser
release:
	@echo "Running GoReleaser..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser not found. Install it from https://goreleaser.com/install/"; \
		exit 1; \
	fi
	@goreleaser release --clean

# Create a snapshot release (local testing)
release-snapshot:
	@echo "Creating snapshot release..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser not found. Install it from https://goreleaser.com/install/"; \
		exit 1; \
	fi
	@goreleaser release --snapshot --clean

# Dry run release (test without publishing)
release-dry-run:
	@echo "Running release dry-run..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser not found. Install it from https://goreleaser.com/install/"; \
		exit 1; \
	fi
	@goreleaser release --skip=publish --clean

# Build Docker image locally
docker:
	@echo "Building Docker image..."
	@docker build -t scmd:latest .

# Check if GoReleaser config is valid
check-goreleaser:
	@echo "Validating GoReleaser configuration..."
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "Error: goreleaser not found. Install it from https://goreleaser.com/install/"; \
		exit 1; \
	fi
	@goreleaser check

# Install GoReleaser
install-goreleaser:
	@echo "Installing GoReleaser..."
	@go install github.com/goreleaser/goreleaser/v2@latest
	@echo "GoReleaser installed to $(shell go env GOPATH)/bin/goreleaser"

# Help
help:
	@echo "scmd Makefile targets:"
	@echo ""
	@echo "Build & Test:"
	@echo "  all               - Run lint, test, and build (default)"
	@echo "  build             - Build the binary"
	@echo "  build-all         - Build for all platforms"
	@echo "  test              - Run all tests with coverage"
	@echo "  test-short        - Run short tests"
	@echo "  test-v            - Run tests with verbose output"
	@echo "  coverage          - Generate coverage HTML report"
	@echo "  lint              - Run linters"
	@echo "  vet               - Run go vet"
	@echo "  fmt               - Format code"
	@echo ""
	@echo "Development:"
	@echo "  dev               - Run in development mode"
	@echo "  install           - Install to /usr/local/bin"
	@echo "  install-go        - Install to GOPATH/bin"
	@echo "  clean             - Clean build artifacts"
	@echo "  deps              - Tidy and verify dependencies"
	@echo ""
	@echo "Release & Distribution:"
	@echo "  tag               - Create a new git tag (make tag VERSION=v1.0.0)"
	@echo "  release           - Run GoReleaser (requires git tag)"
	@echo "  release-snapshot  - Create snapshot release (local testing)"
	@echo "  release-dry-run   - Dry run release (test without publishing)"
	@echo "  completions       - Generate shell completions"
	@echo "  docker            - Build Docker image"
	@echo "  check-goreleaser  - Validate GoReleaser configuration"
	@echo "  install-goreleaser - Install GoReleaser"
	@echo ""
	@echo "Other:"
	@echo "  generate          - Run go generate"
	@echo "  outdated          - Check for outdated dependencies"
	@echo "  audit             - Run security audit"
	@echo "  help              - Show this help message"
