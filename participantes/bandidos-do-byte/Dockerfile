# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Final stage
FROM scratch

# Define build argument for port
ARG PORT
ENV PORT=${PORT}

WORKDIR /

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy training data
COPY --from=builder /app/training ./training

# Expose port (dynamic)
EXPOSE ${PORT}

# Run the application
CMD ["./main"]
