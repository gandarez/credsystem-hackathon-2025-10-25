FROM golang:1.25-bookworm AS build

RUN apt-get update && apt-get -y install \
    git \
    make \
    ca-certificates

WORKDIR /src

COPY . .

RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/ura-ai-linux-amd64 .

FROM scratch AS bin

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /src/build/ura-ai-linux-amd64 /ura-ai-linux-amd64

ENTRYPOINT ["/ura-ai-linux-amd64"]
