#!/bin/bash

set -e

echo "Iniciando testes de integração..."

# Limpar ambiente anterior
docker-compose -f test/integration/docker-compose.test.yml down -v --remove-orphans

# Construir e executar testes
docker-compose -f test/integration/docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from test

# Capturar código de saída
TEST_EXIT_CODE=$?

# Limpar ambiente
docker-compose -f test/integration/docker-compose.test.yml down -v --remove-orphans

# Sair com o código de saída dos testes
exit $TEST_EXIT_CODE
