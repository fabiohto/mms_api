# MMS API - Moving Average Service

O MMS API é um serviço que calcula e disponibiliza médias móveis simples (MMS) para pares de criptomoedas BRL/BTC e BRL/ETH, utilizando dados do Mercado Bitcoin.

## Índice

- [Visão Geral](#visão-geral)
- [Arquitetura](#arquitetura)
- [Tecnologias](#tecnologias)
- [Pré-requisitos](#pré-requisitos)
- [Configuração](#configuração)
- [Executando o Projeto](#executando-o-projeto)
- [Testes](#testes)
- [Monitoramento](#monitoramento)
- [Estrutura do Projeto](#estrutura-do-projeto)

## Visão Geral

O serviço consiste em dois componentes principais:
- **API**: Fornece endpoints REST para consulta de médias móveis
- **Worker**: Executa periodicamente para calcular e atualizar as médias móveis

A atualização dos dados é realizada através de um job agendado que é executado uma vez por dia, consumindo a API do Mercado Bitcoin para obter as informações mais recentes dos pares de criptomoedas.

O sistema calcula três tipos de médias móveis:
- MMS20 (20 períodos)
- MMS50 (50 períodos)
- MMS200 (200 períodos)

## Arquitetura

O projeto segue os princípios da Arquitetura Hexagonal (Ports and Adapters), com:

- Domínio isolado com suas regras de negócio
- Adaptadores para entrada (HTTP API) e saída (Database, External APIs)
- Portas bem definidas através de interfaces
- Injeção de dependências para melhor testabilidade

## Tecnologias

- **Go 1.21**: Linguagem principal do projeto
- **PostgreSQL**: Banco de dados para armazenamento das médias móveis
- **Docker & Docker Compose**: Containerização e orquestração
- **Gin**: Framework web para a API REST
- **Prometheus**: Monitoramento de métricas
- **MailHog**: Servidor SMTP para testes de email
- **Make**: Automação de comandos comuns

## Pré-requisitos

- Go 1.21 ou superior
- Docker e Docker Compose
- Make (opcional, mas recomendado)

## Configuração

1. Clone o repositório
```bash
git clone <repository-url>
cd mms_api
```

2. Copie o arquivo de exemplo de variáveis de ambiente
```bash
cp .env.example .env
```

3. Ajuste as variáveis no arquivo `.env` conforme necessário

## Executando o Projeto

### Usando Make

1. Iniciar todos os serviços
```bash
make up
```

2. Executar apenas a API
```bash
make run-api
```

3. Executar apenas o Worker
```bash
make run-worker
```

4. Parar todos os serviços
```bash
make down
```

### Usando Docker Compose diretamente

1. Construir as imagens
```bash
docker compose build
```

2. Iniciar os serviços
```bash
docker compose up -d
```

3. Verificar logs
```bash
docker compose logs -f
```

## Testes

O projeto inclui testes unitários e de integração.

### Testes Unitários
```bash
make test
```

### Testes de Integração
```bash
make integration-test
```

### Cobertura de Testes
```bash
make coverage
```

## Monitoramento

- **Métricas**: Acessíveis via Prometheus em `http://localhost:9090`
- **Alertas**: Configurados via email, com suporte a diferentes tipos de notificação
- **Logs**: Formato JSON para fácil integração com ferramentas de análise

### Visualização de Alertas no MailHog

O projeto utiliza o MailHog como servidor SMTP para capturar e visualizar emails de alerta, tanto em ambiente de desenvolvimento quanto durante a execução dos testes de integração.

#### Acessando o MailHog

1. Interface Web do MailHog: `http://localhost:8025`
   - Visualize todos os emails enviados
   - Interface intuitiva com preview em tempo real
   - Filtragem e busca de mensagens

#### Cenários de Visualização

1. **Durante Testes de Integração**:
   ```bash
   make integration-test
   ```
   - Os emails de teste serão capturados automaticamente pelo MailHog
   - Acesse a interface web para verificar os alertas gerados pelos testes
   - Os emails são limpos a cada nova execução dos testes

2. **Em Ambiente de Desenvolvimento**:
   - Quando ocorrer um erro real que gere alerta
   - Os emails serão enviados automaticamente para o MailHog
   - Acesse a interface web para analisar os alertas em tempo real
   - Útil para validar o formato e conteúdo dos alertas

#### Tipos de Alertas

Os alertas podem ser visualizados no MailHog incluindo:
- Falhas na coleta de dados
- Erros de processamento
- Problemas de conectividade
- Alertas de performance

## Estrutura do Projeto

```
.
├── cmd/                    # Pontos de entrada da aplicação
│   ├── api/               # Servidor API
│   └── worker/            # Worker para processamento periódico
├── config/                # Configurações da aplicação
├── docker/                # Arquivos Docker e configurações
├── internal/              # Código interno da aplicação
│   ├── adapter/           # Adaptadores (Arquitetura Hexagonal)
│   ├── application/       # Lógica de aplicação
│   └── domain/           # Regras e modelos de domínio
├── pkg/                   # Pacotes reutilizáveis
├── scripts/              # Scripts úteis e migrações
└── test/                 # Testes unitários e de integração
```

## Documentação da API

### Swagger UI

A documentação interativa da API está disponível através do Swagger UI. Para acessá-la:

1. Inicie o serviço:
```bash
make up
```

2. Acesse a documentação em seu navegador:
```
http://localhost:8080/swagger/index.html
```

### Desenvolvimento da Documentação

Para desenvolvedores que precisam atualizar a documentação:

1. A documentação é gerada automaticamente a partir das anotações no código
2. Após fazer alterações nas anotações, regenere a documentação:
```bash
make swagger
```

3. As alterações estarão disponíveis imediatamente no Swagger UI

### Recursos da Documentação

A documentação interativa oferece:
- **Descrição Detalhada**: Todos os endpoints documentados
- **Playground**: Teste as APIs diretamente pela interface
- **Modelos**: Visualize os formatos de request/response
- **Códigos de Status**: Todos os possíveis retornos documentados
- **Parâmetros**: Descrição clara dos parâmetros necessários
- **Exemplos**: Requests e responses de exemplo
- **Autenticação**: Métodos de autenticação suportados (quando aplicável)

## Endpoints

### Consultar MMS por Par
```
GET /api/v1/mms?pair=BRLBTC&from=1620000000&to=1620086400&range=20
```

Parâmetros:
- `pair`: Par de moedas (BRLBTC ou BRLETH)
- `from`: Timestamp Unix de início
- `to`: Timestamp Unix de fim (opcional, default: dia anterior)
- `range`: Período da média móvel (20, 50 ou 200)

