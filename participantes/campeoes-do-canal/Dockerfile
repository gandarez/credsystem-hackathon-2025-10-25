FROM golang:1.25-bookworm AS build

RUN apt-get update && apt-get -y install \
    git \
    make \
    ca-certificates

WORKDIR /src

# Copy everything but defined in docker ignore file
COPY . .

# Build
RUN go mod vendor
RUN make build-linux-amd64

#####################
# Build final image #
#####################
FROM scratch AS bin

# Copy from build
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /src/build/ivr-service-linux-amd64 .

# Specify the container's entrypoint as the action
ENTRYPOINT ["./ivr-service-linux-amd64"]
