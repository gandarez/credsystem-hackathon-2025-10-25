# Bandidos do Byte - Hackathon Solution

Solução para o Hackathon Credsystem & Golang SP 2025.

## Estrutura do Projeto

```
.
├── cmd/
│   └── api/
│       └── main.go           # Entry point da aplicação
├── internal/
│   ├── config/
│   │   └── config.go         # Configurações da aplicação
│   ├── domain/
│   │   └── models.go         # Modelos de domínio
│   ├── handler/
│   │   └── handler.go        # HTTP handlers
│   ├── service/
│   │   └── service.go        # Lógica de negócio
│   └── server/
│       └── server.go         # Configuração do servidor HTTP
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Arquitetura

A aplicação segue os princípios da **Arquitetura Hexagonal** com as seguintes camadas:

- **Domain**: Entidades e modelos de negócio
- **Service**: Lógica de negócio e casos de uso
- **Handler**: Adaptadores HTTP (entrada)
- **Config**: Configurações da aplicação

## Tecnologias Utilizadas

- **Go 1.21**
- **Chi Router**: Router HTTP leve e performático
- **Uber FX**: Framework de injeção de dependências
- **Docker**: Containerização

## Como Executar Localmente

1. Clone o repositório e entre na pasta do projeto:
```bash
cd participantes/bandidos_do_byte
```

2. Copie o arquivo de exemplo de variáveis de ambiente:
```bash
cp .env.example .env
```

3. Configure as variáveis de ambiente no arquivo `.env`

4. Instale as dependências:
```bash
go mod download
```

5. Execute a aplicação:
```bash
go run cmd/api/main.go
```

## Como Executar com Docker

1. Build da imagem:
```bash
docker build -t bandidos-do-byte:latest .
```

2. Execute com docker-compose:
```bash
docker-compose up
```

## Endpoints

### POST /api/find-service
Encontra o serviço mais adequado baseado na intenção do usuário.

**Request:**
```json
{
  "intent": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "service_id": 1,
    "service_name": "Nome do Serviço"
  },
  "error": ""
}
```

### GET /api/healthz
Verifica a saúde do serviço.

**Response:**
```json
{
  "status": "ok"
}
```

## Próximos Passos

- [ ] Implementar integração com OpenRouter API
- [ ] Adicionar lógica de IA para classificação de intenções
- [ ] Carregar dados do arquivo `intents_pre_loaded.csv`
- [ ] Implementar cache de respostas
- [ ] Adicionar testes unitários e de integração
- [ ] Otimizar performance para os limites de recursos (0.5 CPU, 128MB RAM)

## Autores

Bandidos do Byte
