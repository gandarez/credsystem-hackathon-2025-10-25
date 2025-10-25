# Bandidos do Byte - Hackathon Solution

Solução para o Hackathon Credsystem & Golang SP 2025.

## 🎯 Destaques

- ✅ **Arquitetura Hexagonal** completa
- ✅ **OpenRouter + Mistral** para classificação de intenções
- ✅ **TensorFlow** opcional para classificação local

## 🏗️ Estrutura do Projeto

```
.
├── cmd/
│   └── api/
│       └── main.go              # Entry point da aplicação
├── internal/
│   ├── adapters/
│   │   ├── csv_repository.go    # Adapter para dados CSV
│   │   ├── openrouter_client.go # Adapter OpenRouter
│   │   └── tensorflow_classifier.go # Adapter TensorFlow
│   ├── config/
│   │   └── config.go            # Configurações
│   ├── domain/
│   │   ├── intent.go            # Domínio de intents
│   │   └── models.go            # Modelos de domínio
│   ├── handler/
│   │   └── handler.go           # HTTP handlers
│   ├── ports/
│   │   └── ports.go             # Interfaces (portas)
│   ├── service/
│   │   └── service.go           # Lógica de negócio
│   └── server/
│       └── server.go            # Configuração do servidor
├── training/
│   ├── service_intent_model.h5  # Modelo TensorFlow treinado
│   ├── tokenizer.pkl            # Tokenizer para o modelo
│   ├── model_server.py          # Servidor Flask para o modelo
│   └── create_tokenizer.py      # Script para criar tokenizer
├── Dockerfile
├── Dockerfile.tensorflow         # Dockerfile do servidor TF
├── docker-compose.yml
├── CLASSIFIER_GUIDE.md          # Guia detalhado dos classificadores
└── go.mod
```

## 🚀 Classificadores de IA

### OpenRouter Classifier (Padrão)
- Usa API OpenRouter com modelo Mistral-7B
- Alta precisão com LLM
- Classificação baseada em contexto

### TensorFlow Classifier (Opcional)
- Usa algoritmo de similaridade de texto (cosine similarity)
- Rápido e sem dependências externas
- Baseado nos dados de treinamento do CSV
- **Nota**: Não requer modelo .h5 ou servidor Python

## 🔄 Como Trocar o Classificador

Para usar TensorFlow ao invés de OpenRouter, altere no `.env`:
```bash
CLASSIFIER_TYPE=tensorflow
```

## 🛠️ Tecnologias Utilizadas

- **Go 1.21**: Linguagem principal
- **Chi Router**: Router HTTP leve e performático
- **Uber FX**: Framework de injeção de dependências
- **OpenRouter API**: Classificação com Mistral
- **TensorFlow/Keras**: Modelo de ML local
- **Flask**: Servidor para o modelo TensorFlow
- **Docker**: Containerização

## ⚙️ Como Executar

### Pré-requisitos

- Go 1.21+
- Docker (opcional)

### Local

```bash
# 1. Configurar variáveis
export PORT=18020
export OPENROUTER_API_KEY=sua_chave

# 2. Executar
go run cmd/api/main.go
```

### Docker Compose

```bash
docker-compose up -d
```

## 📡 Endpoints

### POST /find-service
Encontra o serviço adequado baseado na intenção.

**Request:**
```json
{
  "intent": "quero abrir uma conta"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "service_id": 1,
    "service_name": "Abertura de Conta"
  }
}
```

### GET /healthz
Verifica a saúde do serviço.

**Response:**
```json
{
  "status": "ok"
}
```

## 🧪 Testes

```bash
curl -X POST http://localhost:18020/find-service \
  -H "Content-Type: application/json" \
  -d '{"intent": "preciso de um empréstimo"}'
```

## 📊 Serviços Disponíveis

O sistema classifica 17 tipos de serviços:

1. Abertura de Conta
2. Empréstimo Pessoal
3. Cartão de Crédito
4. Investimentos
5. Seguros
6. Consórcio
7. Financiamento Imobiliário
8. Financiamento de Veículos
9. Previdência Privada
10. Conta Digital
11. Portabilidade de Salário
12. Renegociação de Dívidas
13. Antecipação de FGTS
14. Crédito Consignado
15. Conta para Empresas
16. Suporte Técnico
0. Contate a URA (fallback)

## 🎓 Arquitetura Hexagonal

```
┌─────────────────────────────────────────┐
│           HTTP Handler (Adapter)         │
└─────────────────┬───────────────────────┘
                  │
         ┌────────▼────────┐
         │  Service (Core)  │
         └────────┬────────┘
                  │
      ┌───────────┴───────────┐
      │                        │
┌─────▼──────┐        ┌───────▼────────┐
│ OpenRouter │   OU   │  TensorFlow    │
│  (Adapter) │        │   (Adapter)    │
└────────────┘        └────────────────┘
```

**Vantagens:**
- ✅ Fácil trocar implementações
- ✅ Testável (mock das interfaces)
- ✅ Desacoplado
- ✅ Escalável

## 🚧 Próximos Passos

- [x] Implementar integração com OpenRouter API
- [x] Adicionar lógica de IA para classificação
- [x] Carregar dados do CSV
- [x] Implementar classificador TensorFlow alternativo
- [ ] Adicionar cache de respostas
- [ ] Testes unitários e de integração

##  Autores

**Bandidos do Byte** - Hackathon Credsystem 2025

