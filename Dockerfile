## Build
FROM golang:1.19-bullseye AS build

WORKDIR /rest-go

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /rest

## Deploy
FROM debian:bullseye

WORKDIR /

COPY --from=build /rest /bin/rest
COPY config.yml.sample /etc/rest-go/config.yml

EXPOSE 3000

ENTRYPOINT ["/bin/rest"]

CMD ["-config", "/etc/rest-go/config.yml"]
