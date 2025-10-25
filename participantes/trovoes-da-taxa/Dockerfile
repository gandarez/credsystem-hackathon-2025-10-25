# Build stage
FROM golang:1.24-alpine AS builder

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

# Criar diretório nlp e copiar data-set.csv
RUN mkdir -p /root/nlp
COPY --from=builder /app/nlp/data-set.csv /root/nlp/data-set.csv

# Expor porta
EXPOSE 18020

# Definir variáveis de ambiente
ENV PORT=18020
ENV INTENTS_CSV_PATH=/root/nlp/data-set.csv

CMD ["./knn-server"]


