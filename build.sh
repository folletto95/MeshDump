#!/bin/sh
set -e

OS=${1:-linux}
ARCH=${2:-amd64}

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

build() {
    os=$1
    arch=$2
    version=$3

    goarch=$arch
    goarm=""
    case "$arch" in
        armhf)
            goarch=arm
            goarm=7
            ;;
        arm64)
            goarch=arm64
            ;;
    esac

    output="MeshDump-${version}-${arch}"
    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi

    echo "Building $os/$arch binary using Docker..."
    docker run --rm -v "$PWD":/src -w /src golang:1.23 \
        sh -c "go mod tidy && \
        GOOS=$os GOARCH=$goarch GOARM=$goarm go build -ldflags '-X meshdump/internal/meshdump.Version=$version' -buildvcs=false -o $output ./cmd/meshdump"

    chmod +x "$output"
    echo "Binary available at $output"
}

if [ "$OS" = "all" ] && { [ "$ARCH" = "all" ] || [ -z "$2" ]; }; then
    build linux amd64 "$new_version"
    build windows amd64 "$new_version"
    build linux armhf "$new_version"
    build linux arm64 "$new_version"
elif [ "$OS" = "all" ]; then
    for os in linux windows; do
        build "$os" "$ARCH" "$new_version"
    done
elif [ "$OS" = "rpi" ]; then
    build linux armhf "$new_version"
    build linux arm64 "$new_version"
else
    build "$OS" "$ARCH" "$new_version"
fi
