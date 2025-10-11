APP_NAME=certfix-agent
CONTAINER=certfix-agent-dev
BUILD_DIR=build

# Build for all supported architectures
build-all: clean
	@echo "Building for all supported architectures..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building for Linux x86_64..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd
	@echo "Building for Linux ARMv7..."
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-armv7 ./cmd
	@echo "All builds completed!"

# Build local (default to x86_64 for compatibility)
build:
	@echo "Building for Linux x86_64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd

# Build specific architectures
build-amd64:
	@echo "Building for Linux x86_64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd

build-arm64:
	@echo "Building for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd

build-armv7:
	@echo "Building for Linux ARMv7..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-armv7 ./cmd

# Build dentro do container
docker-build:
	docker exec $(CONTAINER) make build
	@echo "Build dentro do container concluído!"

# Build all architectures dentro do container
docker-build-all:
	docker exec $(CONTAINER) make build-all
	@echo "Build de todas as arquiteturas dentro do container concluído!"

# Executa o agent dentro do container
docker-run:
	docker exec -it $(CONTAINER) ./build/$(APP_NAME)

# Sobe o ambiente de desenvolvimento
docker-up:
	cd docker && docker-compose up -d --build

# Entra no shell do container
docker-shell:
	docker exec -it $(CONTAINER) /bin/bash

docker-down:
	cd docker && docker-compose down

# Clean build directory
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean completed!"

# Show available targets
help:
	@echo "Available targets:"
	@echo "  build	 - Build for Linux x86_64 (default)"
	@echo "  build-all	 - Build for all supported architectures"
	@echo "  build-amd64   - Build for Linux x86_64"
	@echo "  build-arm64   - Build for Linux ARM64"
	@echo "  build-armv7   - Build for Linux ARMv7"
	@echo "  docker-build  - Build inside container"
	@echo "  docker-build-all - Build all architectures inside container"
	@echo "  docker-run	- Run agent inside container"
	@echo "  docker-up	 - Start development environment"
	@echo "  docker-shell  - Enter container shell"
	@echo "  docker-down   - Stop development environment"
	@echo "  clean	 - Clean build directory"
	@echo "  help	  - Show this help"

.PHONY: build build-all build-amd64 build-arm64 build-armv7 docker-build docker-build-all docker-run docker-up docker-shell docker-down clean help