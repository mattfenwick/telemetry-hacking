#!/usr/bin/env bash

set -xv
set -euo pipefail

./build.sh

java -javaagent:opentelemetry-javaagent.jar \
  -Dotel.metrics.exporter=none \
  -Dotel.service.name=localhost \
  -Dotel.traces.port=16686 \
  -Dotel.traces.exporter=jaeger \
  -jar target/hacking-1.0-SNAPSHOT-jar-with-dependencies.jar \
  localhost


#IMAGE=docker.io/mfenwick100/hacking-java:latest
#
#docker run $IMAGE
