# Cacadores de Corrida - Solução Hackathon

## Arquitetura da Solução

### Abordagem: Agente Único com Validação Determinística

Esta solução utiliza uma arquitetura otimizada para o desafio:

1. **Agente Único (IA)**: Classifica a intenção do cliente usando GPT-4o-mini
2. **Validação Determinística (Código)**: Garante que apenas serviços válidos (1-16) sejam retornados

### Por que essa abordagem?

- ✅ **Eficiência**: 1 chamada de API vs 2 (economia de créditos e tempo)
- ✅ **Performance**: Menor latência = melhor pontuação no ranking
- ✅ **Recursos**: Usa menos memória e CPU (dentro dos limites de 128MB/50% CPU)
- ✅ **Confiabilidade**: Validação em código garante respostas válidas sem custo adicional

## Estrutura do Projeto

```
participantes/cacadores-de-corrida/
├── main.go                 # Servidor HTTP e rotas
├── agent/
│   ├── classifier.go       # Agente IA que classifica intenções
│   └── prompt.go          # System prompt otimizado com exemplos
├── validator/
│   └── validator.go       # Validação determinística dos serviços
├── .env.example           # Exemplo de variáveis de ambiente
├── Dockerfile             # Imagem Docker otimizada
├── docker-compose.yml     # Configuração com limites de recursos
└── go.mod                 # Dependências Go
```

## Como Usar

### 1. Configurar Variáveis de Ambiente

```bash
# Copiar o arquivo de exemplo
cp .env.example .env

# Editar e adicionar sua chave da OpenRouter
# .env
OPENROUTER_API_KEY=sua_chave_aqui
PORT=18020
```

### 2. Executar Localmente (Desenvolvimento)

```bash
# Instalar dependências
go mod download

# Executar
go run main.go
```

### 3. Executar com Docker

```bash
# Build da imagem
docker build -t cacadores-de-corrida .

# Executar com docker-compose
docker-compose up
```

### 4. Testar a API

```bash
# Health check
curl http://localhost:18020/api/healthz

# Classificar intenção
curl -X POST http://localhost:18020/api/find-service \
  -H "Content-Type: application/json" \
  -d '{"intent": "quero mais limite"}'
```

## Endpoints

### POST /api/find-service

Classifica a intenção do cliente e retorna o serviço mais adequado.

**Request:**
```json
{
  "intent": "quero mais limite"
}
```

**Response (Sucesso):**
```json
{
  "success": true,
  "data": {
    "service_id": 6,
    "service_name": "Solicitação de aumento de limite"
  }
}
```

**Response (Erro):**
```json
{
  "success": false,
  "error": "Mensagem de erro"
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

## Estratégia de Prompt Engineering

O system prompt inclui:
- ✅ Lista completa dos 16 serviços
- ✅ Exemplos de classificação (few-shot learning)
- ✅ Instruções claras para não inventar serviços
- ✅ Formato de resposta estruturado
- ✅ Fallback para "Atendimento humano" em caso de dúvida

## Validação em Duas Camadas

1. **IA (Agente)**: Classifica baseado em contexto e exemplos
2. **Código (Validator)**: Verifica se service_id está entre 1-16

## Otimizações para o Ranking

- 🚀 **Velocidade**: 1 chamada de API = menor latência
- 💰 **Custo**: Economia de créditos usando modelo eficiente
- 🎯 **Precisão**: Prompt otimizado com exemplos
- ✅ **Confiabilidade**: Validação garante respostas sempre válidas

## Modelo de IA Utilizado

**openai/gpt-4o-mini**
- Rápido e eficiente
- Ótimo custo-benefício
- Bom desempenho em classificação de texto
- Uso moderado de memória

## Checklist de Conformidade

- ✅ Endpoint `/api/find-service` implementado
- ✅ Endpoint `/api/healthz` implementado
- ✅ Lê variável `OPENROUTER_API_KEY`
- ✅ Lê variável `PORT`
- ✅ Usa apenas os 16 serviços listados
- ✅ Porta 18020 exposta
- ✅ Limites: 50% CPU e 128MB RAM
- ✅ Dockerfile otimizado
- ✅ docker-compose.yml configurado

## Pontuação Esperada

Com essa arquitetura, esperamos:
- Alta taxa de acertos nos 93 testes conhecidos
- Boa generalização para os 80 testes novos
- Tempo de resposta otimizado (< 2000ms por requisição)
- Pontuação final competitiva no ranking

## Autor

Tayson Martins - Hackathon Credsystem & Golang SP 2025
