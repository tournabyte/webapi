FROM golang:1.26 as builder

WORKDIR /usr/src/tournabyte/webapi

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

# copy source code directories to the working directory
COPY ./app/ ./app
COPY ./internal/ ./internal/

# vet the source code for any suspicious code
RUN go vet -v ./...

# execute unit tests to verify source code correctness
RUN go test -v ./...

# build the application
RUN go build -v -o /usr/local/bin/app ./app/start/

# copy application configuration files
WORKDIR /etc/tournabyte
COPY ./appconf.json ./

CMD ["app", "serve"]
