#!/bin/sh
set -e

OS=${1:-linux}
ARCH=${2:-amd64}

build() {
    os=$1
    arch=$2
    output="MeshDump"
    if [ "$os" = "windows" ]; then
        output="MeshDump.exe"
    fi

    version_file="VERSION"
    if [ ! -f "$version_file" ]; then
        echo "0.0" > "$version_file"
    fi
    version=$(cat "$version_file")
    major=${version%%.*}
    minor=${version#*.}
    minor=$((minor + 1))
    new_version="$major.$minor"
    echo "$new_version" > "$version_file"

    echo "Building $os/$arch binary using Docker..."
    docker run --rm -v "$PWD":/src -w /src golang:1.23 \
        sh -c "go mod tidy && \
        GOOS=$os GOARCH=$arch go build -ldflags '-X meshdump/internal/meshdump.Version=$new_version' -buildvcs=false -o $output ./cmd/meshdump"

    echo "Binary available at $output"
}

if [ "$OS" = "all" ]; then
    for os in linux windows; do
        build "$os" "$ARCH"
    done
else
    build "$OS" "$ARCH"
fi
