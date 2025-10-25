# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar go.mod primeiro para cache de dependências
COPY go.mod go.sum* ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar binário do stage anterior
COPY --from=builder /app/main .

# Expor porta
EXPOSE 18020

# Comando para executar
CMD ["./main"]
