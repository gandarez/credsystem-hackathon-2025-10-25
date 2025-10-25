FROM golang:1.25-bookworm AS build

WORKDIR /src

# Install certificates - needed for HTTPS requests
RUN apt-get update && apt-get install -y ca-certificates

# Copy only go.mod and go.sum first for better layer caching
COPY participantes/controladores_do_panico_16/go.* ./
RUN go mod download

# Copy the rest of the source code
COPY participantes/controladores_do_panico_16/ .

# Copy CSV file from repository root
COPY assets/intents_pre_loaded.csv /assets/intents_pre_loaded.csv

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

# Copy CSV file to the container
COPY --from=build /assets/intents_pre_loaded.csv /assets/intents_pre_loaded.csv

# Run the service
ENTRYPOINT ["/ivr-service"]