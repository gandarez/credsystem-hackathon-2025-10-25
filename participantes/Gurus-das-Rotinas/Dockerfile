# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY participantes/Gurus-das-Rotinas/go.mod participantes/Gurus-das-Rotinas/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY participantes/Gurus-das-Rotinas/ .


# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM scratch

# Copy the binary from builder stage
COPY --from=builder /app/main /main

# Copy ca-certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port 18020
EXPOSE 18020

# Run the application
CMD ["/main"]
