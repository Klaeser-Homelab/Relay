
# Relay Voice Server Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build info
BINARY_NAME=relay-voice
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the server
build:
	$(GOBUILD) -o $(BINARY_NAME) -v .

# Build for linux (for Docker)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v .

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run the server
run:
	$(GOBUILD) -o $(BINARY_NAME) -v .
	./$(BINARY_NAME)

# Run with live reload (requires air)
dev:
	air

# Test
test:
	$(GOTEST) -v ./...

# Install air for development
install-air:
	$(GOGET) -u github.com/air-verse/air

# Docker build
docker-build:
	docker build -t relay-voice .

# Docker run
docker-run:
	docker run -p 8080:8080 -e OPENAI_API_KEY=$(OPENAI_API_KEY) relay-voice

# Help
help:
	@echo "Available commands:"
	@echo "  build         - Build the binary"
	@echo "  build-linux   - Build for Linux (Docker)"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  run           - Build and run the server"
	@echo "  dev           - Run with live reload (requires air)"
	@echo "  test          - Run tests"
	@echo "  install-air   - Install air for development"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"

.PHONY: build build-linux clean deps run dev test install-air docker-build docker-run help
