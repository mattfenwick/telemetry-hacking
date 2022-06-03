#!/usr/bin/env bash

set -xv
set -euo pipefail

protoc \
  --go_out=./protobuf \
  --go_opt=paths=source_relative \
  --go-grpc_out=./protobuf \
  --go-grpc_opt=paths=source_relative \
  ./api.proto
