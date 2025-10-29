#!/bin/bash

# Script para exportar variáveis de ambiente e subir o docker-compose

# Carregar variáveis do arquivo config.env
if [ -f "config.env" ]; then
    echo "Carregando variáveis do config.env..."
    export $(grep -v '^#' config.env | xargs)
    echo "Variáveis carregadas:"
    echo "OPENROUTER_API_KEY: ${OPENROUTER_API_KEY:0:20}..."
    echo "OPENROUTER_MODEL: ${OPENROUTER_MODEL}"
    echo "PORT: ${PORT}"
else
    echo "Arquivo config.env não encontrado!"
    exit 1
fi

# Verificar se as variáveis necessárias estão definidas
if [ -z "$OPENROUTER_API_KEY" ]; then
    echo "ERRO: OPENROUTER_API_KEY não está definida!"
    exit 1
fi

echo ""
echo "Subindo docker-compose..."
docker compose up --build -d

echo ""
echo "Verificando status do container..."
docker compose ps

echo ""
echo "Testando endpoint..."
sleep 5
curl -X GET http://localhost:18020/api/healthz
