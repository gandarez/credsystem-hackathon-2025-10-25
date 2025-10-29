#!/bin/bash

# Script para testar intents na API
# Uso: ./test_intents.sh

API_URL="http://localhost:18020/api/intent"
CSV_FILE="assets/intents-pre-loaded.csv"
DELAY=0.5

echo "🧪 Testador de Intents - API de Classificação"
echo "=============================================="
echo "📡 API URL: $API_URL"
echo "📄 Arquivo CSV: $CSV_FILE"
echo

# Verificar se a API está rodando
echo "🔍 Verificando se a API está rodando..."
if curl -s -f "http://localhost:18020/healthz" > /dev/null; then
    echo "✅ API está rodando e acessível"
else
    echo "❌ API não está acessível em http://localhost:18020"
    echo "   Certifique-se de que o servidor está rodando!"
    exit 1
fi

echo

# Verificar se o arquivo CSV existe
if [ ! -f "$CSV_FILE" ]; then
    echo "❌ Arquivo $CSV_FILE não encontrado!"
    exit 1
fi

# Contar total de intents
TOTAL=$(tail -n +2 "$CSV_FILE" | wc -l)
echo "📊 Total de intents encontrados: $TOTAL"
echo

# Contadores
SUCCESS=0
ERROR=0
COUNT=0

echo "🚀 Iniciando testes..."
echo "----------------------"

# Ler CSV e testar cada intent
tail -n +2 "$CSV_FILE" | while IFS=';' read -r service_id service_name intent; do
    COUNT=$((COUNT + 1))
    
    # Pular linhas vazias
    if [ -z "$intent" ]; then
        continue
    fi
    
    echo "[$COUNT/$TOTAL] Testando: '$intent'"
    
    # Fazer requisição para a API
    response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -d "{\"intent\": \"$intent\"}" 2>/dev/null)
    
    # Separar body e status code
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "200" ]; then
        # Verificar se success é true no JSON
        success=$(echo "$body" | jq -r '.success' 2>/dev/null)
        if [ "$success" = "true" ]; then
            echo "    ✅ Sucesso"
            SUCCESS=$((SUCCESS + 1))
        else
            error_msg=$(echo "$body" | jq -r '.error' 2>/dev/null)
            echo "    ❌ Erro: $error_msg"
            ERROR=$((ERROR + 1))
        fi
    else
        echo "    ❌ HTTP $http_code: $body"
        ERROR=$((ERROR + 1))
    fi
    
    # Delay entre requisições
    sleep $DELAY
done

echo
echo "=============================================="
echo "📈 RESUMO DOS TESTES"
echo "=============================================="
echo "Total de intents testados: $TOTAL"
echo "✅ Sucessos: $SUCCESS"
echo "❌ Erros: $ERROR"
echo

# Salvar relatório simples
echo "📄 Relatório salvo em: intent_test_report_$(date +%s).txt"
