#!/usr/bin/env bash

set -xv
set -euo pipefail

IMAGE=docker.io/mfenwick100/telemetry-hacking-java:latest

mvn clean compile assembly:single

docker build -t $IMAGE .
