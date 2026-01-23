#!/usr/bin/env bash

set -euo pipefail

# Project settings
APP_NAME="agc-migration"
OUT_DIR="bin"

if [[ ! -f VERSION ]]; then
  echo "VERSION file not found" >&2
  exit 1
fi

# Strip whitespace/newlines from version string
VERSION="$(tr -d ' \n\r\t' < VERSION)"

if [[ -z "$VERSION" ]]; then
  echo "VERSION file is empty" >&2
  exit 1
fi

# Build settings
export CGO_ENABLED=0
LDFLAGS="-s -w"

# Target platforms (OS/ARCH)
PLATFORMS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

mkdir -p "$OUT_DIR"

for platform in "${PLATFORMS[@]}"; do
  IFS=/ read -r GOOS GOARCH <<< "$platform"

  EXT=""
  if [ "$GOOS" = "windows" ]; then
    EXT=".exe"
  fi

  OUTPUT="$OUT_DIR/${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}${EXT}"

  echo "Building $OUTPUT"
  GOOS="$GOOS" GOARCH="$GOARCH" \
    go build \
      -trimpath \
      -ldflags="$LDFLAGS" \
      -o "$OUTPUT" \
      ./cmd
done

echo "Build complete. Binaries are in $OUT_DIR/"