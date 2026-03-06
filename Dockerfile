FROM golang:1.26

WORKDIR /usr/src/tournabyte/webapi

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY ./app/ ./app
COPY ./internal/ ./internal/
RUN go build -v -o /usr/local/bin/app ./app/start/

# copy application configuration files
WORKDIR /etc/tournabyte
COPY ./appconf.json ./

CMD ["app", "serve"]
