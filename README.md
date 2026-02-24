# eloq-cloud-utils

Shared Go utilities and development infrastructure for EloqCloud services.

## Features

### gRPC Wrapper (`pkg/grpcwrapper`)
- **Singleton Connection Pool**: Reuses physical TCP connections via `singleflight` and `xsync`.
- **Automatic Recovery**: Detects and prunes dead connections (e.g., after server restarts).
- **Production Defaults**: Standardized keepalives, retry policies, and message size limits (64MB/128MB).

## Development Base Image

Used to provide a consistent build environment across all projects.

**Image Tag**: `eloqdata/eloqcloud-dev:go-1.26`

### How to Use

In your project's `Dockerfile`, use it as the builder stage:

```dockerfile
FROM eloqdata/eloqcloud-dev:go-1.26 AS builder

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# All tools (protoc, swag, etc.) are pre-installed
RUN make build
```

The image is automatically built and pushed to Docker Hub via GitHub Actions whenever changes are made to the `development/` directory.

