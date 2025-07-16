#!/bin/sh
set -e

# load environment variables from .env if present
if [ -f .env ]; then
    # shellcheck disable=SC1091
    . ./.env
fi

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

# track built binaries so we can commit them later
built_files=""

# ensure git author identity is set using environment variables when available
GIT_USER_EMAIL=${GIT_USER_EMAIL:-builder@example.com}
GIT_USER_NAME=${GIT_USER_NAME:-MeshDump Builder}

if ! git config user.email >/dev/null; then
    git config user.email "$GIT_USER_EMAIL"
fi
if ! git config user.name >/dev/null; then
    git config user.name "$GIT_USER_NAME"
fi

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
    built_files="$built_files $output"
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

# automatically commit and push the built binaries and version file
if [ -n "$built_files" ]; then
    # shellcheck disable=SC2086 # built_files is intentionally unquoted
    git add "$version_file" $built_files
    git commit -m "Add compiled binaries for version $new_version"
    if git remote | grep -q .; then
        git push
    fi
fi
