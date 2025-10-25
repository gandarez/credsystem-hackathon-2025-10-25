# Etapa de build
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar go.mod e go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build estático (essencial para rodar em scratch)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /app/server ./cmd/server

# Etapa final — imagem mínima
FROM scratch

# Copiar certificados raiz (necessários para HTTPS, se aplicável)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /root/

# Copiar o binário
COPY --from=builder /app/server .

# Copiar assets
COPY --from=builder /app/assets ./assets

EXPOSE 18020

CMD ["./api"]
