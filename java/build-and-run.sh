#!/usr/bin/env bash

set -xv
set -euo pipefail

IMAGE=docker.io/mfenwick100/hacking-java:latest

mvn clean compile assembly:single

java -javaagent:opentelemetry-javaagent.jar \
  -Dotel.metrics.exporter=none \
  -Dotel.service.name=localhost \
  -Dotel.traces.port=16686 \
  -Dotel.traces.exporter=jaeger \
  -jar target/hacking-1.0-SNAPSHOT-jar-with-dependencies.jar

docker build -t $IMAGE .

docker run $IMAGE
