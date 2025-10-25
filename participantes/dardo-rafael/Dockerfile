# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar go.mod primeiro para cache de dependências
COPY go.mod ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o knn-server .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar binário
COPY --from=builder /app/knn-server .

# Copiar CSV de intenções (assumindo que está em ../../assets/)
COPY ../../../assets/intents_pre_loaded.csv /root/assets/intents_pre_loaded.csv

# Expor porta
EXPOSE 18020

# Definir variável de ambiente para o CSV
ENV INTENTS_CSV_PATH=/root/assets/intents_pre_loaded.csv
ENV PORT=18020

CMD ["./knn-server"]


