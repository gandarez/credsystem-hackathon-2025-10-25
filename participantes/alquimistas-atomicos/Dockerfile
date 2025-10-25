FROM golang:1.21-bookworm AS build

RUN apt-get update && apt-get -y install \
    git \
    make \
    ca-certificates

WORKDIR /src

# Copy everything but defined in docker ignore file
COPY . .

# Build
RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o alquimistas-atomicos-linux-amd64 .

#####################
# Build final image #
#####################
FROM scratch AS bin

# Copy from build
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /src/alquimistas-atomicos-linux-amd64 .

# Copy CSV file
COPY --from=build /src/assets/intents_pre_loaded.csv /assets/intents_pre_loaded.csv

# Specify the container's entrypoint as the action
ENTRYPOINT ["./alquimistas-atomicos-linux-amd64"]
