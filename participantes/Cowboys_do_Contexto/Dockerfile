# =========================
# Build stage
# =========================
FROM golang:1.25-bookworm AS build

RUN apt-get update && apt-get -y install \
    git \
    make \
    ca-certificates

WORKDIR /src

# Copy everything except what's in .dockerignore
COPY . .

# Build dependencies and binary
RUN go mod vendor
RUN make build-linux-amd64

# =========================
# Final minimal image
# =========================
FROM scratch AS bin

# Copy system certs and timezone data
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compiled binary
COPY --from=build /src/build/ivr-service-linux-amd64 /ivr-service

# Entrypoint
ENTRYPOINT ["/ivr-service"]
