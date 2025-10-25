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

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o galaxia-dos-gophers-linux-amd64 \
    ./cmd

#####################
# Final image       #
#####################
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=build /src/galaxia-dos-gophers-linux-amd64 .

EXPOSE 8080

ENTRYPOINT ["./galaxia-dos-gophers-linux-amd64"]