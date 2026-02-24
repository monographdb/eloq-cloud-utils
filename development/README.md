# EloqCloud Development Base Image

This directory contains the source and configuration for the EloqCloud standardized development base image. This image provides a consistent, pre-configured environment for building all EloqCloud microservices.

**Image Tag**: `eloqdata/eloqcloud-dev:go-1.26`

## Included Tools

- **Go 1.26**: The official Go compiler and toolchain.
- **Protocol Buffers**: `protoc` compiler (v29.3) and core development headers.
- **protoc-gen-go** (v1.36.11): Official Go code generator for Protocol Buffers.
- **protoc-gen-go-grpc** (v1.5.1): Official Go gRPC plugin for Protocol Buffers.
- **swag** (v1.16.4): Swagger 2.0 documentation generator for Go.
- **Utility Tools**: `git`, `make`, `curl`, `ca-certificates`.

## Usage Instructions

### Building the Image

Use the provided `Makefile` to manage the build process:

```bash
# Build for the local platform
make build

# Build for multi-arch (linux/amd64 and linux/arm64)
make build-multiarch

# Build and push to Docker Hub (requires credentials)
make build-push
```

### Local Testing

Verify the tools are correctly installed inside the image:
```bash
make test
```

## Using in Your Projects

### Update Your Dockerfile

Replace the existing `FROM` instruction in your service's `Dockerfile` to use this base image:

```dockerfile
# Use the EloqCloud development base image as the builder
FROM eloqdata/eloqcloud-dev:go-1.26 AS builder
```

### Full Builder Example

```dockerfile
FROM eloqdata/eloqcloud-dev:go-1.26 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# Copy Go Modules manifests and download dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Generate Proto files if necessary
RUN protoc -I. -I/usr/include \
    --go_out=./pkg --go-grpc_out=./pkg \
    --go-grpc_opt=require_unimplemented_servers=false \
    --go_opt=paths=source_relative \
    --go-grpc_opt=paths=source_relative \
    proto/*.proto

# Build the application
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -a -o app cmd/main.go

# Final stage: minimal runtime environment
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/app .
USER 65532:65532
ENTRYPOINT ["/app"]
```

## Maintenance

### Updating Tool Versions

To update the version of a specific tool, modify the `go install` commands in the `Dockerfile`:

```dockerfile
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1 && \
    go install github.com/swaggo/swag/cmd/swag@v1.16.4
```

### List of Dependent Projects

The following projects are configured to use this base image for their CI/CD and production builds:
- `eloq-shelf`
- `eloq-operator`
- `eloqcloud-proxy`
- `eloqcloud-control-plane`

## License

Standard EloqCloud licensing applies.
