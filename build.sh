#!/bin/sh
set -e

OS=${1:-linux}
ARCH=${2:-amd64}

output="MeshDump"
if [ "$OS" = "windows" ]; then
    output="MeshDump.exe"
fi

echo "Building $OS/$ARCH binary using Docker..."
docker run --rm -v "$PWD":/src -w /src golang:1.23 \
    sh -c "go mod tidy && \
    GOOS=$OS GOARCH=$ARCH go build -buildvcs=false -o $output ./cmd/meshdump"

echo "Binary available at $output"
