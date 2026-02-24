# syntax=docker/dockerfile:1
# EloqCloud Development Base Image
# This image contains common build dependencies for EloqCloud projects
ARG GO_VERSION=1.26
FROM golang:${GO_VERSION} AS base

# Install build essentials and protobuf compiler
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y --no-install-recommends \
    git \
    make \
    curl \
    unzip \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install specific version of protoc
ARG PROTOC_VERSION=29.3
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local && \
    rm protoc-${PROTOC_VERSION}-linux-x86_64.zip

# Configure Go environment for better build performance
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV CGO_ENABLED=0

# Install Go tools with version pinning for consistency
# Use cache mount to speed up rebuilds
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1 && \
    go install github.com/swaggo/swag/cmd/swag@v1.16.6

# Verify installations
RUN protoc --version && \
    go version && \
    protoc-gen-go --version && \
    protoc-gen-go-grpc --version && \
    swag --version

# Set working directory
WORKDIR /workspace

# Default command
CMD ["/bin/bash"]
