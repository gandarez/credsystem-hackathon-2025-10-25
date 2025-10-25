FROM golang:1.25-bookworm AS build

RUN apt-get update && apt-get -y install \
    git \
    ca-certificates

WORKDIR /src

COPY . .

RUN go mod download

# Build the binary to a predictable location in the build stage
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app ./main.go

FROM scratch AS bin

# Copy CA certificates into the same path inside the scratch image
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# Copy the built binary from the build stage
COPY --from=build /app /app

EXPOSE 8080

ENTRYPOINT ["/app"]
