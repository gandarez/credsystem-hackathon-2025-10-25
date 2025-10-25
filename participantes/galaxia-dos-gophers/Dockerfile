# Build stage
FROM golang:1.23-bookworm AS build

RUN apt-get update && apt-get -y install \
    git \
    make \
    ca-certificates

WORKDIR /src

# Copy go files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o galaxia-dos-gophers-linux-amd64 \
    ./cmd

#####################
# Final image       #
#####################
FROM scratch

# Set the working directory
WORKDIR /app

# Copy certificates and timezone data from the build stage
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=build /src/galaxia-dos-gophers-linux-amd64 .

# Expose the application port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["./galaxia-dos-gophers-linux-amd64"]