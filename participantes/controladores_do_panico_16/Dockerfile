FROM golang:1.25-bookworm AS build

WORKDIR /src

# Install certificates - needed for HTTPS requests
RUN apt-get update && apt-get install -y ca-certificates

# Copy only go.mod and go.sum first for better layer caching
COPY go.* ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build
ARG CGO_ENABLED
ARG GOOS
ARG GOARCH
RUN go build -o ivr-service

# Create minimal runtime image
FROM scratch AS runtime

# Copy SSL certificates and timezone data from build stage
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary from build stage
COPY --from=build /src/ivr-service /ivr-service

# Run the service
ENTRYPOINT ["/ivr-service"]