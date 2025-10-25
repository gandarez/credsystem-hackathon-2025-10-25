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
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .


# Expose port 18020
EXPOSE 18020

# Run the application
CMD ["./main"]
