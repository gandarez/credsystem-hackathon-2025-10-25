.PHONY: run build test clean docker-build docker-run

# Variáveis
BINARY_NAME=knn-server
DOCKER_IMAGE=knn-classifier
PORT?=18020

# Build da aplicação
build:
	@echo "Building..."
	go build -o $(BINARY_NAME) .

# Executar localmente
run:
	@echo "Running..."
	go run .

# Executar testes
test:
	@echo "Running tests..."
	go test -v ./...

# Limpar binários
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)

# Build Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

# Executar Docker
docker-run:
	@echo "Running Docker container..."
	docker run -p $(PORT):18020 \
		-e OPENROUTER_API_KEY="$(OPENROUTER_API_KEY)" \
		$(DOCKER_IMAGE)

# Executar com hot reload (requer air: go install github.com/cosmtrek/air@latest)
dev:
	@echo "Running with hot reload..."
	air

# Formatar código
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint
lint:
	@echo "Linting..."
	golangci-lint run

# Verificar dependências
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Help
help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make dev          - Run with hot reload"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter"
	@echo "  make deps         - Download dependencies"


