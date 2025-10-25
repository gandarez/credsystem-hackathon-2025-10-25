# Etapa de build
FROM golang:1.25.3-alpine AS builder

WORKDIR /app

# Copiar go.mod
COPY go.mod ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build estático (essencial para rodar em scratch)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /app/api ./cmd/api

# Etapa final — imagem mínima
FROM scratch

# Copiar certificados raiz (necessários para HTTPS, se aplicável)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /root/

# Copiar o binário
COPY --from=builder /app/api .

# Copiar assets
COPY --from=builder /app/assets ./assets

EXPOSE 18020

CMD ["./api"]
