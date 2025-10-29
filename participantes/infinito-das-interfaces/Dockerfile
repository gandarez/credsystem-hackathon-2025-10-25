FROM golang:1.24-bookworm AS build

RUN apt-get update && apt-get -y install \
    git \
    make \
    ca-certificates

WORKDIR /src
COPY . .

RUN go mod vendor
RUN make build-linux-amd64

FROM scratch AS bin

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /src/build/cs_ura_intent-linux-amd64 ./cs_ura_intent

ENTRYPOINT ["./cs_ura_intent"]