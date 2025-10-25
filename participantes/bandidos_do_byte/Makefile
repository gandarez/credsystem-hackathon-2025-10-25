.PHONY: build run docker-build docker-run test clean

# Build the application
build:
	go build -o bin/api cmd/api/main.go

# Run the application locally
run:
	go run cmd/api/main.go

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	docker build -t bandidos-do-byte:latest .

# Run with Docker Compose
docker-run:
	docker-compose up

# Run with Docker Compose in detached mode
docker-run-detached:
	docker-compose up -d

# Stop Docker Compose
docker-stop:
	docker-compose down

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build               - Build the application"
	@echo "  run                 - Run the application locally"
	@echo "  deps                - Download and tidy dependencies"
	@echo "  docker-build        - Build Docker image"
	@echo "  docker-run          - Run with Docker Compose"
	@echo "  docker-run-detached - Run with Docker Compose in background"
	@echo "  docker-stop         - Stop Docker Compose"
	@echo "  test                - Run tests"
	@echo "  clean               - Clean build artifacts"
	@echo "  fmt                 - Format code"
	@echo "  lint                - Run linter"
