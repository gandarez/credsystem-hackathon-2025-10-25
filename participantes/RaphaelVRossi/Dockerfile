FROM golang:1.25-bookworm AS build

WORKDIR /src

COPY service/. .

RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-app .


FROM scratch AS bin

COPY --from=build go-app .

ENTRYPOINT ["./go-app"]
