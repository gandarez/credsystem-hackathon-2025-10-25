#!/bin/bash

# Script de teste com 5 requisições

echo "======================================"
echo "TEST SCRIPT - 5 Requisitions"
echo "======================================"
echo ""

# 1. Parar servidor anterior
echo "[1/4] Stopping previous server..."
sudo lsof -ti:8080 | xargs sudo kill -9 2>/dev/null
sleep 1

# 2. Iniciar servidor
echo "[2/4] Starting server..."
cd /home/mauricio/Desktop/credsystem-hackathon-2025-10-25/participantes/controladores_do_panico_16
OPENROUTER_API_KEY='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' go run main.go > test_server.log 2>&1 &
SERVER_PID=$!
echo "Server started with PID: $SERVER_PID"

# 3. Aguardar servidor iniciar
echo "[3/4] Waiting for server to start..."
sleep 4

# 4. Fazer 5 requisições de teste
echo "[4/4] Running 5 test requests..."
echo ""

# Array com as 5 intenções de teste
intents=(
    "quero saber meu limite"
    "perdi meu cartão"
    "segunda via da fatura"
    "quero aumentar meu limite"
    "não lembro minha senha"
)

# Array para armazenar resultados
results='{"test_date":"'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'","total_requests":5,"requests":[],"summary":{}}'

success_count=0
fail_count=0
total_time=0

echo "Starting requests..."
echo ""

for i in "${!intents[@]}"; do
    request_num=$((i + 1))
    intent="${intents[$i]}"
    
    echo "Request $request_num: \"$intent\""
    
    # Medir tempo de resposta
    start_time=$(date +%s%3N)
    
    # Fazer requisição
    response=$(curl -s -X POST http://localhost:8080/api/find-service \
        -H "Content-Type: application/json" \
        -d "{\"intent\": \"$intent\"}" \
        -w "\n%{http_code}\n%{time_total}")
    
    end_time=$(date +%s%3N)
    
    # Extrair dados da resposta
    http_code=$(echo "$response" | tail -n 2 | head -n 1)
    time_total=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -2)
    
    # Calcular tempo em ms
    time_ms=$(echo "$time_total * 1000" | bc | cut -d'.' -f1)
    total_time=$((total_time + time_ms))
    
    # Verificar sucesso
    if [ "$http_code" = "200" ]; then
        success=$(echo "$body" | jq -r '.success // false')
        if [ "$success" = "true" ]; then
            success_count=$((success_count + 1))
            status="SUCCESS"
        else
            fail_count=$((fail_count + 1))
            status="FAILED"
        fi
    else
        fail_count=$((fail_count + 1))
        status="ERROR"
    fi
    
    echo "  Status: $status | HTTP: $http_code | Time: ${time_ms}ms"
    echo "  Response: $body"
    echo ""
    
    # Adicionar ao JSON (formato simplificado para bash)
    request_json=$(cat <<EOF
{
  "request_number": $request_num,
  "intent": "$intent",
  "http_code": $http_code,
  "response_time_ms": $time_ms,
  "status": "$status",
  "response_body": $body
}
EOF
)
    
    # Salvar em arquivo temporário
    echo "$request_json" >> /tmp/test_results_$request_num.json
done

# 5. Calcular estatísticas e pontuação
echo "======================================"
echo "TEST RESULTS SUMMARY"
echo "======================================"
echo ""
echo "Total Requests: 5"
echo "Successes: $success_count"
echo "Failures: $fail_count"

avg_time=$((total_time / 5))
echo "Average Response Time: ${avg_time}ms"
echo ""

# Calcular pontuação conforme README
# Score = (Total_Sucessos × 10.0) - (Total_Falhas × 50.0) - (Tempo_Médio_ms × 0.01)
score_success=$(echo "$success_count * 10.0" | bc)
score_fail=$(echo "$fail_count * 50.0" | bc)
score_time=$(echo "$avg_time * 0.01" | bc)
total_score=$(echo "$score_success - $score_fail - $score_time" | bc)

echo "SCORE CALCULATION:"
echo "  Successes: $success_count × 10.0 = $score_success points"
echo "  Failures:  $fail_count × 50.0 = -$score_fail points"
echo "  Avg Time:  ${avg_time}ms × 0.01 = -$score_time points"
echo "  ----------------------------------------"
echo "  TOTAL SCORE: $total_score points"
echo ""

# 6. Criar JSON final com todos os resultados
echo "Generating results JSON file..."

cat > test_results_5req.json <<EOF
{
  "test_info": {
    "date": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
    "total_requests": 5,
    "test_type": "manual_5_requests"
  },
  "requests": [
EOF

# Adicionar cada requisição
for i in {1..5}; do
    if [ -f "/tmp/test_results_$i.json" ]; then
        cat "/tmp/test_results_$i.json" >> test_results_5req.json
        if [ $i -lt 5 ]; then
            echo "," >> test_results_5req.json
        fi
    fi
done

# Completar o JSON
cat >> test_results_5req.json <<EOF

  ],
  "summary": {
    "total_successes": $success_count,
    "total_failures": $fail_count,
    "average_response_time_ms": $avg_time,
    "total_time_ms": $total_time
  },
  "scoring": {
    "formula": "(successes × 10.0) - (failures × 50.0) - (avg_time_ms × 0.01)",
    "success_points": $score_success,
    "failure_penalty": -$score_fail,
    "time_penalty": -$score_time,
    "total_score": $total_score
  }
}
EOF

# Limpar arquivos temporários
rm -f /tmp/test_results_*.json

echo "Results saved to: test_results_5req.json"
echo ""
echo "Server logs saved to: test_server.log"
echo ""
echo "To stop the server: sudo kill $SERVER_PID"
echo ""
echo "======================================"
