BINARY_NAME=droidlink
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS=-s -w -X main.Version=$(VERSION)
INSTALL_DIR?=$(HOME)/.local/bin

.PHONY: all build install run test lint clean tidy help docker-build docker-run

all: tidy test build

## Build the application
build:
	@echo "Building..."
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .

## Install the binary into a user-level bin directory
install:
	@echo "Installing $(BINARY_NAME) into $(INSTALL_DIR)..."
	install -d "$(INSTALL_DIR)"
	go build -trimpath -ldflags "$(LDFLAGS)" -o "$(INSTALL_DIR)/$(BINARY_NAME)" .

## Run the application (pass ARGS="...")
run:
	@echo "Running..."
	go run . $(ARGS)

## Run tests
test:
	@echo "Running tests..."
	go test -v ./...

## Run golangci-lint
lint:
	golangci-lint run ./...

## Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/

## Tidy dependencies
tidy:
	@echo "Tidying modules..."
	go mod tidy

## Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build --build-arg VERSION=$(VERSION) -t $(BINARY_NAME) .

## Run Docker image
docker-run:
	@echo "Running Docker image..."
	docker run --rm $(BINARY_NAME)

## Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  install      - Install the binary into ~/.local/bin (override INSTALL_DIR)"
	@echo "  run          - Run the application (pass ARGS=\"...\")"
	@echo "  test         - Run tests"
	@echo "  lint         - Run golangci-lint"
	@echo "  clean        - Clean build artifacts"
	@echo "  tidy         - Tidy Go modules"
	@echo "  docker-build - Build the Docker image"
	@echo "  docker-run   - Run the Docker image"
	@echo "  all          - Run tidy, test, and build"
