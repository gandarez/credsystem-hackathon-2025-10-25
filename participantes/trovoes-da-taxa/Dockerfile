#####################
# Build stage       #
#####################
FROM golang:1.25-bookworm AS build

RUN apt-get update && apt-get -y install \
    git \
    make \
    ca-certificates

WORKDIR /app

# Copy everything
COPY . .

# Download dependencies and build
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o knn-server .

#####################
# Build final image #
#####################
FROM scratch AS bin

# Copy from build
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /app/knn-server .
COPY --from=build /app/nlp/data-set.csv /nlp/data-set.csv

# Definir variáveis de ambiente
ENV PORT=18020
ENV INTENTS_CSV_PATH=/nlp/data-set.csv

# Specify the container's entrypoint as the action
ENTRYPOINT ["./knn-server"]


