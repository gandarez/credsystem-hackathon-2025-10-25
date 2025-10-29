# Alquimistas Atômicos

Serviço de classificação de intenções para URA da Credsystem usando IA.

## 🚀 Funcionalidades

- **POST /api/find-service**: Classifica intenções usando IA (OpenRouter)
- **GET /api/healthz**: Verifica saúde do serviço
- **Otimizado para recursos limitados**: 128MB RAM, 0.5 CPU

## 🛠️ Tecnologias

- **Go 1.21**: Linguagem principal
- **Gorilla Mux**: Roteamento HTTP
- **OpenRouter API**: Integração com IA
- **Docker**: Containerização
- **Alpine Linux**: Imagem base minimalista

## 📦 Estrutura do Projeto

```
├── main.go                 # Servidor HTTP principal
├── go.mod                  # Dependências Go
├── client/                 # Cliente OpenRouter
│   └── openrouter.go
├── handlers/               # Handlers dos endpoints
│   └── service.go
├── models/                 # Estruturas de dados
│   └── response.go
├── Dockerfile              # Imagem Docker
├── docker-compose.yml      # Orquestração
└── Makefile               # Comandos de desenvolvimento
```

## 🏃‍♂️ Como Executar

### 🔑 **Configuração Inicial (OBRIGATÓRIO)**

```bash
# 1. Configurar chave OpenRouter
./setup.sh sk-or-v1-sua_chave_aqui

# 2. Ou editar manualmente o arquivo config.env
# OPENROUTER_API_KEY=sk-or-v1-sua_chave_aqui
# PORT=8080
```

### **Executar Localmente**

```bash
# Instalar dependências
make deps

# Executar (carrega automaticamente config.env)
make run
```

### **Executar com Docker**

```bash
# Build da imagem
make docker-build

# Executar com Docker Compose (carrega config.env automaticamente)
make docker-run

# Ou manualmente com variáveis carregadas
export $(cat config.env | xargs) && docker-compose up --build
```

## 🔧 Variáveis de Ambiente

- `OPENROUTER_API_KEY`: Chave da API OpenRouter (obrigatória)
- `PORT`: Porta do serviço (padrão: 8080)

## 📊 Endpoints

### POST /api/find-service

**Request:**
```json
{
  "intent": "quero aumentar meu limite"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "service_id": 6,
    "service_name": "Solicitação de aumento de limite"
  }
}
```

### GET /api/healthz

**Response:**
```json
{
  "status": "ok"
}
```

## 🧠 Estratégia de IA

1. **Classificação por IA**: Usa OpenRouter com modelo Mistral 7B (gratuito e eficiente)
2. **Fallback CSV**: Busca exata nos dados pré-carregados
3. **Atendimento Humano**: Último recurso para casos não classificados

## 🐳 Docker

- **Imagem base**: Scratch (ultra minimalista - apenas o binário)
- **Recursos limitados**: 128MB RAM, 0.5 CPU
- **Porta**: 18020 (conforme especificação)
- **Multi-stage build**: Otimizado para produção
- **Tamanho**: Imagem ultra compacta (~10MB)

## 📈 Performance

- **Timeout**: 25s para requisições IA
- **Fallback rápido**: < 1ms para busca no CSV
- **Memória otimizada**: Imagem Alpine + binário estático
- **Concorrência**: Suporte a múltiplas requisições simultâneas

---

**Desenvolvido por**: Alquimistas Atômicos 🧪⚛️
