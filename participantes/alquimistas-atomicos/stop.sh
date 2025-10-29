#!/bin/bash

# Script para parar o docker-compose

echo "Parando docker-compose..."
docker compose down

echo ""
echo "Verificando se os containers foram parados..."
docker compose ps
