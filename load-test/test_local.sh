#!/usr/bin/env bash

# Teste Local - Sem Docker
# Roda o serviço diretamente com Go e testa

set -e

PARTICIPANT="controladores_do_panico_16"
PARTICIPANT_DIR="../participantes/$PARTICIPANT"
RESULTS_DIR="$PARTICIPANT_DIR/results"
OUTPUT_FILE="$RESULTS_DIR/test_local.txt"

# Criar diretório de resultados se não existir
if [ ! -d "$RESULTS_DIR" ]; then
    mkdir -p "$RESULTS_DIR"
fi

echo "========================================"
echo "  Teste Local - 1 Requisição (Sem Docker)"
echo "========================================"
echo ""

# Carregar .env se existir e OPENROUTER_API_KEY não estiver definida
if [ -z "$OPENROUTER_API_KEY" ] && [ -f "$PARTICIPANT_DIR/.env" ]; then
    echo "📂 Carregando OPENROUTER_API_KEY do arquivo .env..."
    export $(grep -v '^#' "$PARTICIPANT_DIR/.env" | xargs)
fi

# Verificar se a API key está configurada
if [ -z "$OPENROUTER_API_KEY" ]; then
    echo "❌ ERRO: OPENROUTER_API_KEY não está configurada!"
    echo "Execute: export OPENROUTER_API_KEY='sua_chave_aqui'"
    echo "Ou crie o arquivo: $PARTICIPANT_DIR/.env"
    exit 1
fi

echo "✅ OPENROUTER_API_KEY encontrada"

# Compilar o serviço
echo ""
echo "🔨 Compilando serviço..."
cd "$PARTICIPANT_DIR"
go build -o ivr-service . 2>&1
if [ $? -ne 0 ]; then
    echo "❌ ERRO: Falha na compilação"
    exit 1
fi
echo "✅ Compilação concluída"

# Iniciar o serviço em background
echo ""
echo "🚀 Iniciando serviço na porta 8080..."
PORT=8080 ./ivr-service > /dev/null 2>&1 &
SERVICE_PID=$!

# Aguardar o serviço ficar pronto
echo ""
echo "⏳ Aguardando serviço iniciar..."
max_attempts=10
attempt=1
success=1

sleep 2  # Aguardar um pouco antes de começar

while [ $success -ne 0 ] && [ $max_attempts -ge $attempt ]; do
    if curl -f -s --max-time 2 http://localhost:8080/api/healthz > /dev/null 2>&1; then
        success=0
        echo "✅ Serviço está respondendo!"
    else
        echo "   Tentativa $attempt de $max_attempts..."
        sleep 1
        ((attempt++))
    fi
done

if [ $success -ne 0 ]; then
    echo ""
    echo "❌ ERRO: Serviço não respondeu após $max_attempts tentativas"
    kill $SERVICE_PID 2>/dev/null || true
    exit 1
fi

# Executar UMA requisição de teste
echo ""
echo "========================================="
echo "  Executando Teste com 1 Intenção"
echo "========================================="
echo ""

TEST_INTENT="Quanto tem disponível para usar"
EXPECTED_SERVICE_ID=1
EXPECTED_SERVICE_NAME="Consulta Limite / Vencimento do cartão / Melhor dia de compra"

echo "📝 Intent: \"$TEST_INTENT\""
echo "🎯 Esperado: service_id=$EXPECTED_SERVICE_ID"
echo ""

# Medir tempo de início
start_time=$(date +%s%3N)

# Fazer requisição
response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d "{\"intent\": \"$TEST_INTENT\"}" \
    http://localhost:8080/api/find-service)

# Medir tempo de fim
end_time=$(date +%s%3N)
elapsed=$((end_time - start_time))

# Separar corpo da resposta e status code
http_body=$(echo "$response" | head -n -1)
http_code=$(echo "$response" | tail -n 1)

echo "========================================="
echo "  Resultado"
echo "========================================="
echo ""
echo "⏱️  Tempo de Resposta: ${elapsed}ms"
echo "📊 HTTP Status: $http_code"
echo ""
echo "📄 Response Body:"
echo "$http_body" | jq '.' 2>/dev/null || echo "$http_body"
echo ""

# Validar resultado
test_success=0
test_failed=0

# A API sempre retorna 200, verificar o campo success
if [ "$http_code" -eq 200 ]; then
    success_flag=$(echo "$http_body" | jq -r '.success' 2>/dev/null)
    
    if [ "$success_flag" == "true" ]; then
        service_id=$(echo "$http_body" | jq -r '.data.service_id' 2>/dev/null)
        service_name=$(echo "$http_body" | jq -r '.data.service_name' 2>/dev/null)
        
        if [ "$service_id" == "$EXPECTED_SERVICE_ID" ]; then
            echo "✅ SUCESSO!"
            echo "   service_id: $service_id (correto)"
            echo "   service_name: $service_name"
            test_success=1
        else
            echo "❌ FALHA!"
            echo "   Esperado: service_id=$EXPECTED_SERVICE_ID"
            echo "   Recebido: service_id=$service_id"
            test_failed=1
        fi
    else
        echo "❌ ERRO na resposta"
        error_msg=$(echo "$http_body" | jq -r '.error' 2>/dev/null)
        echo "   Mensagem: $error_msg"
        test_failed=1
    fi
else
    echo "❌ ERRO: HTTP $http_code (inesperado - API deve sempre retornar 200)"
    error_msg=$(echo "$http_body" | jq -r '.error' 2>/dev/null)
    echo "   Mensagem: $error_msg"
    test_failed=1
fi

# Calcular Score conforme README
# Score = (Sucessos × 10.0) - (Falhas × 50.0) - (Tempo_Médio_ms × 0.01)
score=$(echo "scale=2; ($test_success * 10.0) - ($test_failed * 50.0) - ($elapsed * 0.01)" | bc)

echo ""
echo "========================================="
echo "  Pontuação"
echo "========================================="
echo ""
echo "Fórmula: Score = (Sucessos × 10) - (Falhas × 50) - (Tempo_ms × 0.01)"
echo ""
echo "Sucessos: $test_success"
echo "Falhas: $test_failed"
echo "Tempo: ${elapsed}ms"
echo ""
echo "🏆 SCORE: $score pontos"
echo ""

# Salvar resultado em arquivo TXT
{
    echo "========================================"
    echo "  TESTE LOCAL - 1 REQUISIÇÃO (SEM DOCKER)"
    echo "========================================"
    echo ""
    echo "Data/Hora: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "Modelo: openai/gpt-4o-mini"
    echo ""
    echo "----------------------------------------"
    echo "  TESTE"
    echo "----------------------------------------"
    echo "Intent: \"$TEST_INTENT\""
    echo "Service ID Esperado: $EXPECTED_SERVICE_ID"
    echo "Service Name Esperado: $EXPECTED_SERVICE_NAME"
    echo ""
    echo "----------------------------------------"
    echo "  RESULTADO"
    echo "----------------------------------------"
    echo "HTTP Status: $http_code"
    echo "Tempo de Resposta: ${elapsed}ms"
    echo ""
    if [ "$http_code" -eq 200 ]; then
        echo "Service ID Recebido: $service_id"
        echo "Service Name Recebido: $service_name"
        echo ""
        if [ $test_success -eq 1 ]; then
            echo "Status: ✅ SUCESSO"
        else
            echo "Status: ❌ FALHA"
        fi
    else
        echo "Erro: $error_msg"
        echo "Status: ❌ ERRO"
    fi
    echo ""
    echo "----------------------------------------"
    echo "  PONTUAÇÃO"
    echo "----------------------------------------"
    echo "Fórmula: Score = (Sucessos × 10) - (Falhas × 50) - (Tempo_ms × 0.01)"
    echo ""
    echo "Sucessos: $test_success"
    echo "Falhas: $test_failed"
    echo "Tempo: ${elapsed}ms"
    echo ""
    echo "🏆 SCORE: $score pontos"
    echo ""
    echo "----------------------------------------"
    echo "  ESTIMATIVA DE CUSTO"
    echo "----------------------------------------"
    echo "Modelo usado: openai/gpt-4o-mini"
    echo "Custo estimado desta requisição: ~\$0.0001 - \$0.0005"
    echo ""
    echo "💡 Para verificar consumo real:"
    echo "   python ../../utils/check_limit_openrouter.py"
    echo "   Ou acesse: https://openrouter.ai/activity"
    echo ""
    echo "📊 Estimativa para teste completo:"
    echo "   - Teste 93: 93 requisições × ~\$0.0003 = ~\$0.03"
    echo "   - Teste 80: 80 requisições × ~\$0.0003 = ~\$0.02"
    echo "   - Total estimado: ~\$0.05 (dentro do limite de \$3.00)"
    echo ""
    echo "========================================"
} > "$OUTPUT_FILE"

# Parar o serviço
echo ""
echo "🛑 Parando serviço..."
kill $SERVICE_PID 2>/dev/null || true
wait $SERVICE_PID 2>/dev/null || true

echo ""
echo "========================================="
echo "  Estimativa de Custo"
echo "========================================="
echo ""
echo "Modelo usado: openai/gpt-4o-mini"
echo "Custo estimado por requisição: ~\$0.0001 - \$0.0005"
echo ""
echo "💡 Para verificar consumo real:"
echo "   python ../utils/check_limit_openrouter.py"
echo "   Ou acesse: https://openrouter.ai/activity"
echo ""
echo "📊 Estimativa para teste completo:"
echo "   - Teste 93: 93 requisições × ~\$0.0003 = ~\$0.03"
echo "   - Teste 80: 80 requisições × ~\$0.0003 = ~\$0.02"
echo "   - Total estimado: ~\$0.05 (bem dentro do limite de \$3.00)"
echo ""
echo "========================================="
echo ""
echo "📄 Resultado salvo em: $OUTPUT_FILE"
echo ""

cd - > /dev/null
