# Variables
BINARY_NAME=linkmcp
BUILD_DIR=bin
DOCKER_IMAGE=wst-link-mcp:latest
CONTAINER_NAME=wst-link-mcp

# Go binary path (uses system go or custom path)
GO=$(shell which go 2>/dev/null || echo "/usr/local/go/bin/go")

.PHONY: all help build run stop clean test docker-build docker-run docker-stop

all: build

## help: Display available make commands
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build the Go binary into bin/linkmcp"
	@echo "  run           Run the application locally (requires LINKEDIN_ACCESS_TOKEN)"
	@echo "  stop          Stop any running linkmcp local processes or Docker container"
	@echo "  clean         Clean build directory and test artifacts"
	@echo "  test          Run all unit tests"
	@echo "  docker-build  Build the Docker image"
	@echo "  docker-run    Run the server inside Docker using .env file"
	@echo "  docker-stop   Stop the running Docker container"

## build: Build the Go binary
build:
	@echo "Building Go binaries..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/linkmcp/main.go
	$(GO) build -ldflags="-w -s" -o $(BUILD_DIR)/token-refresher ./cmd/token-refresher/main.go


## run: Run the application locally
run: build
	@echo "Starting linkmcp..."
	@if [ -f .env ]; then \
		set -a; source .env; set +a; \
	fi; \
	./$(BUILD_DIR)/$(BINARY_NAME)

## stop: Stop running process or container
stop: docker-stop
	@echo "Stopping local linkmcp process..."
	@pkill -f $(BUILD_DIR)/$(BINARY_NAME) 2>/dev/null || true

## clean: Remove build artifacts
clean:
	@echo "Cleaning build directory and artifacts..."
	@rm -rf $(BUILD_DIR) coverage.out

## test: Run unit tests
test:
	@echo "Running unit tests..."
	$(GO) test -v ./...

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .

## docker-run: Run Docker container with .env file
docker-run:
	@echo "Running Docker container $(CONTAINER_NAME)..."
	@if [ -f .env ]; then \
		docker run -i --rm --name $(CONTAINER_NAME) --env-file .env $(DOCKER_IMAGE); \
	else \
		docker run -i --rm --name $(CONTAINER_NAME) -e LINKEDIN_ACCESS_TOKEN=$${LINKEDIN_ACCESS_TOKEN} $(DOCKER_IMAGE); \
	fi

## docker-stop: Stop Docker container
docker-stop:
	@echo "Stopping Docker container $(CONTAINER_NAME)..."
	@docker stop $(CONTAINER_NAME) 2>/dev/null || true
