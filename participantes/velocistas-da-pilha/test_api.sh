#!/bin/bash

# Script de teste para validar a API

API_URL="http://localhost:18020"

echo "🧪 Iniciando testes da API Velocistas da Pilha"
echo "================================================"

# Teste 1: Health Check
echo ""
echo "1️⃣ Testando /api/healthz..."
HEALTH=$(curl -s $API_URL/api/healthz)
echo "Resposta: $HEALTH"

if echo "$HEALTH" | grep -q '"status":"ok"'; then
    echo "✅ Health check passou!"
else
    echo "❌ Health check falhou!"
    exit 1
fi

# Teste 2: Intenções do CSV (match exato)
echo ""
echo "2️⃣ Testando intenções conhecidas (match exato)..."

test_intent() {
    INTENT=$1
    EXPECTED_ID=$2
    EXPECTED_NAME=$3
    
    RESPONSE=$(curl -s -X POST $API_URL/api/find-service \
        -H "Content-Type: application/json" \
        -d "{\"intent\": \"$INTENT\"}")
    
    SERVICE_ID=$(echo $RESPONSE | grep -o '"service_id":[0-9]*' | grep -o '[0-9]*')
    
    if [ "$SERVICE_ID" = "$EXPECTED_ID" ]; then
        echo "✅ '$INTENT' → ID $SERVICE_ID (correto)"
    else
        echo "❌ '$INTENT' → ID $SERVICE_ID (esperado: $EXPECTED_ID)"
    fi
}

# Testar algumas intenções do CSV
test_intent "Quanto tem disponível para usar" 1 "Consulta Limite"
test_intent "Preciso de uma segunda via do boleto" 2 "Segunda via de boleto"
test_intent "Quero cancelar meu cartão" 7 "Cancelamento de cartão"
test_intent "Perdi meu cartão" 11 "Perda e roubo"
test_intent "Esqueci minha senha" 10 "Esqueceu senha"

# Teste 3: Intenções similares (usando LLM)
echo ""
echo "3️⃣ Testando intenções similares (LLM)..."

test_intent "Qual o meu limite disponível?" 1 "Consulta Limite"
test_intent "Gostaria de aumentar meu limite" 6 "Solicitação de aumento"
test_intent "Meu cartão foi roubado" 11 "Perda e roubo"
test_intent "Quero falar com um atendente" 15 "Atendimento humano"

# Teste 4: Performance
echo ""
echo "4️⃣ Testando performance (10 requisições)..."

START=$(date +%s%3N)
for i in {1..10}; do
    curl -s -X POST $API_URL/api/find-service \
        -H "Content-Type: application/json" \
        -d '{"intent": "Quanto tem disponível para usar"}' > /dev/null
done
END=$(date +%s%3N)

DURATION=$((END - START))
AVG=$((DURATION / 10))

echo "Tempo total: ${DURATION}ms"
echo "Tempo médio: ${AVG}ms por requisição"

if [ $AVG -lt 50 ]; then
    echo "✅ Performance excelente! (<50ms)"
elif [ $AVG -lt 200 ]; then
    echo "✅ Performance boa! (<200ms)"
else
    echo "⚠️  Performance pode melhorar (>${AVG}ms)"
fi

echo ""
echo "================================================"
echo "✅ Testes concluídos!"