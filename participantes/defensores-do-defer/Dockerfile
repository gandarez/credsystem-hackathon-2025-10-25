FROM golang:1.25-bookworm AS build


WORKDIR /app

# Copy everything but defined in docker ignore file
COPY . .

# Build
RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o defendoresDefer ./cmd/api/main.go

#####################
# Build final image #
#####################
FROM scratch

# Copy certificates and timezone from build stage (necessário para TLS e timezone)
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
# Copia o binário gerado para a localização raiz da imagem
COPY --from=build /app/defendoresDefer /defendoresDefer

EXPOSE 8080

ENTRYPOINT ["/defendoresDefer"]