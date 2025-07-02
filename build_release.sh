#!/bin/bash

set -e

APP_NAME="safrochaind"
VERSION="v1.0.0"

echo "🛠 Building binaries for $VERSION..."

# Docker image with GCC for Linux + CGO
IMAGE="messense/rust-musl-cross:x86_64-musl"

docker run --rm -v $(pwd):/volume -w /volume \
  -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=1 \
  -e CC=x86_64-linux-musl-gcc \
  golang:1.23.9 \
  sh -c "apt update && apt install -y musl-tools && go build -o ${APP_NAME}_linux_amd64 ./cmd/safrochaind"

tar -czvf ${VERSION}_linux_amd64.tar.gz ${APP_NAME}_linux_amd64
rm ${APP_NAME}_linux_amd64

# macOS AMD64 (host build)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o ${APP_NAME}_darwin_amd64 ./cmd/safrochaind
tar -czvf ${VERSION}_darwin_amd64.tar.gz ${APP_NAME}_darwin_amd64
rm ${APP_NAME}_darwin_amd64

# macOS ARM64 (host build)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o ${APP_NAME}_darwin_arm64 ./cmd/safrochaind
tar -czvf ${VERSION}_darwin_arm64.tar.gz ${APP_NAME}_darwin_arm64
rm ${APP_NAME}_darwin_arm64

echo "✅ Build done. Files:"
ls -lh ${VERSION}_*.tar.gz