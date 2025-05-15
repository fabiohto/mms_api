# Variáveis
APP_NAME=mms_api
VERSION?=1.0.0
GOCMD=go
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
DOCKER_COMPOSE=docker-compose
DOCKER_COMPOSE_FILE=docker-compose.yml

# Cores para output
GREEN=\033[0;32m
NC=\033[0m # No Color

.PHONY: all test clean docker-build help deps lint integration-test up down logs ps

all: deps test integration-test

help: ## Mostra esta ajuda
	@awk 'BEGIN {FS = ":.*##"; printf "\nUso:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

clean: ## Remove containers, volumes e imagens
	@printf "$(GREEN)Limpando ambiente Docker...$(NC)\n"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans --rmi local

deps: ## Download das dependências
	@printf "$(GREEN)Baixando dependências...$(NC)\n"
	$(GOMOD) download
	$(GOMOD) tidy

lint: ## Executa linter em um container
	@printf "$(GREEN)Executando linter...$(NC)\n"
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint run

test: ## Executa testes unitários em um container
	@printf "$(GREEN)Executando testes unitários...$(NC)\n"
	docker run --rm -v $(PWD):/app -w /app golang:1.21-alpine go test -v ./test/unit/...

integration-test: ## Executa testes de integração com docker-compose
	@printf "$(GREEN)Executando testes de integração...$(NC)\n"
	$(DOCKER_COMPOSE) -f test/integration/docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from test
	$(DOCKER_COMPOSE) -f test/integration/docker-compose.test.yml down -v

build: ## Build das imagens Docker
	@printf "$(GREEN)Construindo imagens Docker...$(NC)\n"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) build

up: ## Inicia os containers
	@printf "$(GREEN)Iniciando containers...$(NC)\n"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d

down: ## Para os containers
	@printf "$(GREEN)Parando containers...$(NC)\n"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) down

ps: ## Lista containers em execução
	@printf "$(GREEN)Listando containers...$(NC)\n"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) ps

logs: ## Mostra logs dos containers
	@printf "$(GREEN)Mostrando logs...$(NC)\n"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) logs -f

migrate: ## Executa migrações do banco de dados
	@printf "$(GREEN)Executando migrações...$(NC)\n"
	docker-compose -f $(DOCKER_COMPOSE_FILE) run --rm api go run $(MIGRATIONS_DIR)/initial_load.go

run-api: ## Executa apenas a API
	@printf "$(GREEN)Executando API...$(NC)\n"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d api

run-worker: ## Executa apenas o Worker
	@printf "$(GREEN)Executando Worker...$(NC)\n"
	$(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_FILE) up -d worker

mock: ## Gera mocks para testes
	@printf "$(GREEN)Gerando mocks...$(NC)\n"
	mockgen -source=internal/application/port/out/mms_repository.go -destination=test/unit/mock/mock_repository.go
	mockgen -source=internal/application/port/out/candle_api.go -destination=test/unit/mock/mock_candle_api.go

coverage: ## Gera relatório de cobertura de testes em um container
	@printf "$(GREEN)Gerando relatório de cobertura...$(NC)\n"
	docker run --rm -v $(PWD):/app -w /app golang:1.21-alpine /bin/sh -c \
		"go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html"

bench: ## Executa benchmarks em um container
	@printf "$(GREEN)Executando benchmarks...$(NC)\n"
	docker run --rm -v $(PWD):/app -w /app golang:1.21-alpine go test -bench=. -benchmem ./...

check: lint test ## Executa verificações (lint + testes)

ci: deps lint test integration-test ## Pipeline de CI completa

# Targets adicionais para Docker
restart: down up ## Reinicia todos os containers

prune: ## Remove todos os recursos Docker não utilizados
	@printf "$(GREEN)Removendo recursos Docker não utilizados...$(NC)\n"
	docker system prune -f

# Target padrão
.DEFAULT_GOAL := help
