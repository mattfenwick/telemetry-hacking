#!/usr/bin/env bash

set -xv
set -euo pipefail

IMAGE=docker.io/mfenwick100/telemetry-hacking:latest

pushd ..
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o ./docker/main \
    ./cmd/main.go
popd

docker build -t $IMAGE .
