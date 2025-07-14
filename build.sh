#!/bin/sh
set -e

echo "Building Windows binary using Docker..."
docker run --rm -v "$PWD":/src -w /src golang:1.20 \
    bash -c 'go mod tidy && \
    GOOS=windows GOARCH=amd64 go build -buildvcs=false -o MeshDump.exe ./cmd/meshdump'
echo "Binary available at MeshDump.exe"

