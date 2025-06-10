# Relay Server Makefile

.PHONY: build test clean help

# Default target
all: build

# Build the relay binary
build:
	@echo "🔨 Building Relay..."
	go build -o relay

# Run the test suite
test: build
	@echo "🧪 Running Relay Test Suite..."
	cd test && go run *.go

# Build test runner
test-build:
	@echo "🔨 Building test runner..."
	cd test && go build -o test_runner *.go

# Run quick smoke test
smoke-test: build
	@echo "💨 Running smoke test..."
	./relay --help || (echo "❌ Relay binary failed to run" && exit 1)
	@echo "✅ Smoke test passed"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f relay
	rm -f test/test_runner
	rm -f test/*.exe

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Run linters
lint:
	@echo "🔍 Running linters..."
	go vet ./...

# Full pipeline: clean, deps, format, lint, build, test
full: clean deps fmt lint build test

# Development mode - watch for changes and rebuild
dev:
	@echo "👀 Development mode (run 'make build' after changes)"
	@echo "Relay binary: ./relay"
	@echo "Test command: make test"

# Create a test project for manual testing
setup-test-project:
	@echo "🏗️ Setting up test project..."
	@if [ ! -d "/Users/reed/Code/Personal/RelayTest" ]; then \
		mkdir -p /Users/reed/Code/Personal/RelayTest; \
		cd /Users/reed/Code/Personal/RelayTest && git init; \
		echo "# RelayTest\n\nThis is a test repository for Relay Server." > README.md; \
		cd /Users/reed/Code/Personal/RelayTest && git add README.md; \
		cd /Users/reed/Code/Personal/RelayTest && git commit -m "Initial commit"; \
		echo "✅ Created RelayTest repository"; \
	else \
		echo "✅ RelayTest repository already exists"; \
	fi

# Run integration test with real RelayTest repo
integration-test: build setup-test-project
	@echo "🔗 Running integration test..."
	./relay add -p "/Users/reed/Code/Personal/RelayTest" || true
	./relay list
	./relay open RelayTest || true
	./relay status || true
	./relay remove RelayTest || true
	@echo "✅ Integration test completed"

# Performance test
perf-test: build
	@echo "⚡ Running performance test..."
	@echo "Testing project operations..."
	time ./relay add -p "/Users/reed/Code/Personal/RelayTest" || true
	time ./relay list
	time ./relay open RelayTest || true
	time ./relay status || true
	time ./relay remove RelayTest || true

# CI/CD commands
ci-test:
	@echo "🚀 Running CI-style tests..."
	@export CI_MODE=true && cd test && go run *.go

ci-setup:
	@echo "🔧 Setting up CI-like environment..."
	@mkdir -p /tmp/RelayTest
	@cd /tmp/RelayTest && git init && git config user.name "CI Test" && git config user.email "ci@test.com"
	@echo "# CI Test Repo" > /tmp/RelayTest/README.md
	@cd /tmp/RelayTest && git add README.md && git commit -m "Initial commit"
	@echo "✅ CI environment ready at /tmp/RelayTest"

pre-commit:
	@echo "🔍 Running pre-commit checks..."
	@$(MAKE) fmt
	@$(MAKE) lint
	@$(MAKE) build
	@$(MAKE) test
	@echo "✅ All pre-commit checks passed"

# Help
help:
	@echo "Relay Server Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build              Build the relay binary"
	@echo "  test               Run the full test suite"
	@echo "  smoke-test         Quick test to verify binary works"
	@echo "  clean              Remove build artifacts"
	@echo "  deps               Install Go dependencies"
	@echo "  fmt                Format Go code"
	@echo "  lint               Run code linters"
	@echo "  full               Run complete pipeline"
	@echo "  dev                Development mode info"
	@echo "  setup-test-project Create RelayTest repository"
	@echo "  integration-test   Manual integration test"
	@echo "  perf-test          Performance testing"
	@echo "  ci-test            Run tests in CI mode"
	@echo "  ci-setup           Set up CI-like environment"
	@echo "  pre-commit         Run all pre-commit checks"
	@echo "  help               Show this help message"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build         # Build relay binary"
	@echo "  make test          # Run all tests"
	@echo "  make full          # Complete build and test"
	@echo "  make ci-test       # Test in CI mode"
	@echo "  make pre-commit    # Pre-commit validation"