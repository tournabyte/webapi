FROM golang:1.26

WORKDIR /usr/src/tournabyte/webapi

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY ./app/ ./internal ./
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]
