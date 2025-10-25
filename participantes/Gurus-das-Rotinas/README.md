# Gurus das Rotinas - Intent Classification Service

This service implements intent classification for URA (IVR assistant) using OpenRouter's embeddings API.

## Features

- **Intent Classification**: Uses OpenRouter's `text-embedding-3-small` model to classify user intents
- **Cosine Similarity**: Finds the most similar service based on embedding similarity
- **Confidence Threshold**: Returns "no match" if confidence is below 75%
- **Precomputed Embeddings**: Loads embeddings from JSON for faster startup
- **Fallback Generation**: Automatically generates embeddings if JSON file doesn't exist

## API Endpoints

### POST /api/find-service
Original endpoint format for the hackathon.

**Request:**
```json
{
  "intent": "quero cancelar meu cartão"
}
```

**Response (Success):**
```json
{
  "success": true,
  "data": {
    "service_id": 7,
    "service_name": "Cancelamento de cartão"
  },
  "error": ""
}
```

**Response (Error):**
```json
{
  "success": false,
  "data": {
    "service_id": 0,
    "service_name": ""
  },
  "error": "Could not determine service from intent"
}
```

### POST /api/classify
New classification endpoint with confidence scores.

**Request:**
```json
{
  "text": "quero cancelar meu cartão"
}
```

**Response:**
```json
{
  "service_id": 7,
  "service_name": "Cancelamento de cartão",
  "confidence": 0.88
}
```

### GET /api/healthz
Verifica a saúde do serviço. Retorna 200 OK se o serviço estiver funcionando corretamente.

**Response:**
```json
{
  "status": "ok"
}
```

## Setup

1. **Set your OpenRouter API key:**
   ```bash
   export OPENROUTER_API_KEY="your_api_key_here"
   ```

2. **Generate embeddings (first time only):**
   ```bash
   cd cmd/generate-embeddings
   go run main.go $OPENROUTER_API_KEY
   cd ../..
   ```

3. **Run the service:**
   ```bash
   docker-compose up --build
   ```

## How it Works

1. **Data Loading**: Loads service examples from CSV file (`../../assets/intents_pre_loaded.csv`)
2. **Embedding Generation**: Uses OpenRouter's embeddings API to generate vector representations
3. **Similarity Matching**: When a new intent comes in:
   - Generates embedding for the input text
   - Calculates cosine similarity with all stored embeddings
   - Returns the service with highest similarity above threshold (75%)
4. **Caching**: Saves embeddings to `service_embeddings.json` for faster subsequent startups

## Services Supported

The service supports 16 predefined services:
1. Consulta Limite / Vencimento do cartão / Melhor dia de compra
2. Segunda via de boleto de acordo
3. Segunda via de Fatura
4. Status de Entrega do Cartão
5. Status de cartão
6. Solicitação de aumento de limite
7. Cancelamento de cartão
8. Telefones de seguradoras
9. Desbloqueio de Cartão
10. Esqueceu senha / Troca de senha
11. Perda e roubo
12. Consulta do Saldo
13. Pagamento de contas
14. Reclamações
15. Atendimento humano
16. Token de proposta

## Example Usage

```bash
# Test the classification endpoint
curl -X POST http://localhost:18020/api/classify \
  -H "Content-Type: application/json" \
  -d '{"text": "quero cancelar meu cartão"}'

# Test the original find-service endpoint
curl -X POST http://localhost:18020/api/find-service \
  -H "Content-Type: application/json" \
  -d '{"intent": "quero cancelar meu cartão"}'
```

## Configuration

- **Port**: 18020 (configurable via `PORT` environment variable)
- **Confidence Threshold**: 75% (hardcoded, can be made configurable)
- **Embedding Model**: `text-embedding-3-small` (cost-effective choice)
- **Resource Limits**: 50% CPU, 128MB RAM (as per hackathon requirements)
