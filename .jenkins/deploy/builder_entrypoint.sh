#!/usr/bin/env bash

set -eo pipefail

function main() {
  local OUTPUT_DIR
  OUTPUT_DIR='dist'

  go mod tidy

  rm -rf "$OUTPUT_DIR"
  mkdir -p "$OUTPUT_DIR"

  CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=$ARCH \
    GOEXPERIMENT=systemcrypto \
    go build \
    -o "$OUTPUT_DIR/conjur" \
    ./cmd/conjur/main.go

  # Ensure the binary is compiled with FIPS enabled
  if ! go tool nm "$OUTPUT_DIR/conjur" | grep 'openssl_FIPS_mode' >/dev/null; then
    echo "FIPS mode not enabled in the binary"
    exit 1
  fi
}

main "$@"
