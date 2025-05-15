FROM golang:1.21-alpine

WORKDIR /app

# Instalar dependências de build
RUN apk add --no-cache gcc musl-dev

# Copiar arquivos do projeto
COPY go.mod go.sum ./
RUN go mod download

# Copiar o resto do código
COPY . .

# Comando para executar os testes
CMD ["go", "test", "-v", "./test/integration/..."]
