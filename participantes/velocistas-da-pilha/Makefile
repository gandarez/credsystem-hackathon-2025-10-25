.PHONY: help run test build docker-build docker-push docker-run clean

# Variáveis
IMAGE_NAME=rogeriogregorio/velocistas-da-pilha
TAG=latest

help: ## Mostra este help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Executa localmente
	@echo "🚀 Executando API..."
	go run cmd/api/main.go

test: ## Testa o CSV loader
	@echo "🧪 Testando loader..."
	go run cmd/test_csv/main.go

test-api: ## Testa a API (requer API rodando)
	@echo "🧪 Testando API..."
	chmod +x test_api.sh
	./test_api.sh

build: ## Build do binário
	@echo "🔨 Building..."
	go build -o main cmd/api/main.go

docker-build: ## Build da imagem Docker
	@echo "🐳 Building Docker image..."
	docker build -t $(IMAGE_NAME):$(TAG) .

docker-push: docker-build ## Push da imagem para registry
	@echo "📤 Pushing to registry..."
	docker push $(IMAGE_NAME):$(TAG)

docker-run: ## Executa com Docker Compose
	@echo "🐳 Starting with Docker Compose..."
	docker-compose up -d
	@echo "✅ Container rodando!"
	@echo "📊 Logs: docker-compose logs -f"

docker-stop: ## Para containers
	@echo "🛑 Stopping containers..."
	docker-compose down

docker-logs: ## Mostra logs
	docker-compose logs -f

clean: ## Limpa arquivos temporários
	@echo "🧹 Cleaning..."
	rm -f main
	go clean

check: ## Valida estrutura do projeto
	@echo "✅ Verificando estrutura..."
	@test -f assets/intents_pre_loaded.csv && echo "  ✅ CSV encontrado" || echo "  ❌ CSV não encontrado"
	@test -f Dockerfile && echo "  ✅ Dockerfile encontrado" || echo "  ❌ Dockerfile não encontrado"
	@test -f docker-compose.yml && echo "  ✅ docker-compose.yml encontrado" || echo "  ❌ docker-compose.yml não encontrado"
	@test -f go.mod && echo "  ✅ go.mod encontrado" || echo "  ❌ go.mod não encontrado"
	@test -d internal/classifier && echo "  ✅ classifier encontrado" || echo "  ❌ classifier não encontrado"
	@test -d internal/storage && echo "  ✅ storage encontrado" || echo "  ❌ storage não encontrado"
	@echo "✅ Validação completa!"