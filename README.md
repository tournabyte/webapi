# Package `github.com/tournabyte/webapi`

## Overview

This go module provides a RESTful API for the data models powering the Tournabyte platform.

## Getting Started

### Prerequisites

- Go `1.26.x` (see `go.mod`)
- GNU `make`

### Build

From the repository root:

```bash
make deps
make build
```

This creates the binary at `./bin/tbyte-webapi`.

### Install

You can either:

- use the built binary directly (`./bin/tbyte-webapi`), or
- install with Go:

```bash
go install github.com/tournabyte/webapi@latest
```

### Configure

The application reads `webapi.json` from these locations:

1. `/etc/tournabyte`
2. `$HOME/.local/tournabyte`
3. the directory provided by `--config` (defaults to `.`)

Create a `webapi.json` file with this structure:

```json
{
  "serve": {
    "port": 8080,
    "security": {
      "useTLS": false,
      "certificateFile": "/path/to/tls-cert.pem",
      "keychainFile": "/path/to/tls-key.pem"
    },
    "sessions": {
      "signingAlgorithm": "HS256",
      "signingKeyFile": "/path/to/session-signing-key",
      "accessTokenTTL": "15m",
      "refreshTokenTTL": "72h",
      "tokenIssuer": "example.com",
      "tokenSubject": "Tournabyte API"
    }
  },
  "mongodb": {
    "hosts": ["mongodb01.example.com", "mongodb02.example.com"],
    "username": "/path/to/mongodb-username",
    "password": "/path/to/mongodb-password"
  },
  "minio": {
    "endpoint": "minio.example.com:9000",
    "accessKey": "/path/to/minio-access-key",
    "secretKey": "/path/to/minio-secret-key"
  },
  "log": {
    "destinations": ["stdout"],
    "prefix": "webapi",
    "flags": 3
  }
}
```

Important:

- several fields are file paths that are resolved at startup (for example signing key and database/object store credentials),
- secret/key files must use restrictive permissions (for example `0600`, and `0400` for TLS cert/key files where required).

You can validate configuration before running:

```bash
go run . check --config /path/to/config-directory --load
```

### Operate

Show command help:

```bash
go run . --help
```

Start the API server:

```bash
go run . server --config /path/to/config-directory
```

Override the configured port at runtime:

```bash
go run . server --config /path/to/config-directory --port 8080
```

### Development checks

```bash
make test
make vet
make build
```
