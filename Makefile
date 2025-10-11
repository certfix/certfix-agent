APP_NAME=certfix-agent
CONTAINER=certfix-agent-dev
BUILD_DIR=build

# Build for all supported architectures
build-all: clean
	@echo "Building for all supported architectures..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building for Linux x86_64..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd
	@echo "Building for Linux ARMv7..."
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-armv7 ./cmd
	@echo "All builds completed!"

# Build local (default to x86_64 for compatibility)
build:
	@echo "Building for Linux x86_64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd

# Build for development (native platform)
build-dev:
	@echo "Building for development (native platform)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME)-dev ./cmd

# Build specific architectures
build-amd64:
	@echo "Building for Linux x86_64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd

build-arm64:
	@echo "Building for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd

build-armv7:
	@echo "Building for Linux ARMv7..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux-armv7 ./cmd

# Docker build targets
docker-build:
	docker exec $(CONTAINER) make build
	@echo "Build inside container completed!"

docker-build-all:
	docker exec $(CONTAINER) make build-all
	@echo "Build all architectures inside container completed!"

docker-build-dev:
	docker exec $(CONTAINER) make build-dev
	@echo "Development build inside container completed!"

# Run targets
run: build
	@echo "Running agent..."
	./$(BUILD_DIR)/$(APP_NAME)

docker-run: docker-build-dev
	@echo "Running agent inside container..."
	docker exec -it $(CONTAINER) ./build/$(APP_NAME)-dev

# Test targets
test:
	@echo "Running tests..."
	go test ./...

docker-test:
	docker exec $(CONTAINER) make test

# Docker environment management
docker-up:
	cd docker && docker-compose up -d --build

docker-down:
	cd docker && docker-compose down

docker-shell:
	docker exec -it $(CONTAINER) /bin/bash

docker-logs:
	docker logs -f $(CONTAINER)

# Utility targets
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean completed!"

mod-tidy:
	@echo "Tidying Go modules..."
	go mod tidy

mod-download:
	@echo "Downloading Go modules..."
	go mod download

# Release preparation
prepare-release: clean build-all
	@echo "Preparing release artifacts..."
	@ls -la $(BUILD_DIR)/
	@echo "Release preparation completed!"

# Show available targets
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build		 - Build for Linux x86_64 (default)"
	@echo "  build-dev	 - Build for development (native platform)"
	@echo "  build-all	 - Build for all supported architectures"
	@echo "  build-amd64   - Build for Linux x86_64"
	@echo "  build-arm64   - Build for Linux ARM64"
	@echo "  build-armv7   - Build for Linux ARMv7"
	@echo ""
	@echo "Docker build targets:"
	@echo "  docker-build	 - Build inside container"
	@echo "  docker-build-dev - Build for development inside container"
	@echo "  docker-build-all - Build all architectures inside container"
	@echo ""
	@echo "Run targets:"
	@echo "  run		 - Build and run agent locally"
	@echo "  docker-run  - Build and run agent inside container"
	@echo ""
	@echo "Test targets:"
	@echo "  test		- Run tests locally"
	@echo "  docker-test - Run tests inside container"
	@echo ""
	@echo "Docker environment:"
	@echo "  docker-up	- Start development environment"
	@echo "  docker-down  - Stop development environment"
	@echo "  docker-shell - Enter container shell"
	@echo "  docker-logs  - View container logs"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean		   - Clean build directory"
	@echo "  mod-tidy		- Tidy Go modules"
	@echo "  mod-download	- Download Go modules"
	@echo "  prepare-release - Prepare release artifacts"
	@echo "  help			- Show this help"

.PHONY: build build-dev build-all build-amd64 build-arm64 build-armv7 \
		docker-build docker-build-dev docker-build-all \
		run docker-run test docker-test \
		docker-up docker-down docker-shell docker-logs \
		clean mod-tidy mod-download prepare-release help