.PHONY: build run test docker-build docker-run clean

# Build da aplicação
build:
	go build -o alquimistas-atomicos .

# Executar localmente
run:
	@if [ -f config.env ]; then \
		export $$(grep -v '^#' config.env | xargs) && go run .; \
	else \
		echo "⚠️  Arquivo config.env não encontrado. Criando exemplo..."; \
		echo "OPENROUTER_API_KEY=sua_chave_aqui" > config.env; \
		echo "PORT=8080" >> config.env; \
		echo "✅ Arquivo config.env criado. Edite com sua chave OpenRouter e execute novamente."; \
	fi

# Testes
test:
	go test ./...

# Build da imagem Docker
docker-build:
	docker build -t alquimistas-atomicos/alquimistas-atomicos:latest .

# Executar com Docker Compose
docker-run:
	@if [ -f config.env ]; then \
		export $$(grep -v '^#' config.env | xargs) && docker-compose up --build; \
	else \
		echo "⚠️  Arquivo config.env não encontrado. Execute ./setup.sh primeiro."; \
		exit 1; \
	fi

# Limpar arquivos temporários
clean:
	rm -f alquimistas-atomicos
	go clean

# Instalar dependências
deps:
	go mod tidy
	go mod download
