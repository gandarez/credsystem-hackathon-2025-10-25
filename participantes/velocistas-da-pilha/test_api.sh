#!/bin/bash

API_URL="http://localhost:18020"

echo "🧪 Iniciando testes da API Velocistas da Pilha"
echo "================================================"

# Função de teste de intenção com medição de tempo
test_intent() {
    INTENT=$1
    EXPECTED_ID=$2
    EXPECTED_NAME=$3

    START=$(date +%s%3N)
    RESPONSE=$(curl -s -X POST $API_URL/api/find-service \
        -H "Content-Type: application/json" \
        -d "{\"intent\": \"$INTENT\"}")
    END=$(date +%s%3N)
    DURATION=$((END - START))

    SERVICE_ID=$(echo $RESPONSE | grep -o '"service_id":[0-9]*' | grep -o '[0-9]*')

    if [ "$SERVICE_ID" = "$EXPECTED_ID" ]; then
        echo "✅ '$INTENT' → ID $SERVICE_ID (correto) [$DURATION ms]"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo "❌ '$INTENT' → ID $SERVICE_ID (esperado: $EXPECTED_ID) [$DURATION ms]"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi

    TOTAL_TIME=$((TOTAL_TIME + DURATION))
}

# -----------------------------
# 1️⃣ Health Check
# -----------------------------
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

# -----------------------------
# 2️⃣ Testar intenções conhecidas
# -----------------------------
CSV_FILE="./assets/intents_pre_loaded.csv"

echo ""
echo "2️⃣ Testando todas as intenções conhecidas..."

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_TIME=0
COUNT=0

while IFS=";" read -r SERVICE_ID SERVICE_NAME INTENT; do
    test_intent "$INTENT" "$SERVICE_ID" "$SERVICE_NAME"
    COUNT=$((COUNT + 1))
done < <(tail -n +2 "$CSV_FILE")

if [ $COUNT -gt 0 ]; then
    AVG_TIME=$((TOTAL_TIME / COUNT))
else
    AVG_TIME=0
fi

echo "⏱️ Tempo total: ${TOTAL_TIME} ms"
echo "⏱️ Tempo médio: ${AVG_TIME} ms por requisição"
echo "✅ Passaram: $PASS_COUNT"
echo "❌ Falharam: $FAIL_COUNT"

# -----------------------------
# 3️⃣ Testar intenções similares (LLM)
# -----------------------------
EXTRA_CSV="./assets/extra-intents.csv"

echo ""
echo "3️⃣ Testando todas as intenções similares..."

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_TIME=0
COUNT=0

while IFS=";" read -r EXPECTED_ID EXPECTED_NAME INTENT; do
    test_intent "$INTENT" "$EXPECTED_ID" "$EXPECTED_NAME"
    COUNT=$((COUNT + 1))
done < <(tail -n +2 "$EXTRA_CSV")

if [ $COUNT -gt 0 ]; then
    AVG_TIME=$((TOTAL_TIME / COUNT))
else
    AVG_TIME=0
fi

echo "⏱️ Tempo total: ${TOTAL_TIME} ms"
echo "⏱️ Tempo médio: ${AVG_TIME} ms por requisição"
echo "✅ Passaram: $PASS_COUNT"
echo "❌ Falharam: $FAIL_COUNT"

echo ""
echo "================================================"
echo "✅ Testes concluídos!"
