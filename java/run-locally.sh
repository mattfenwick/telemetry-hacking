#!/usr/bin/env bash

set -xv
set -euo pipefail

./build.sh

# see: https://github.com/open-telemetry/opentelemetry-java/blob/main/sdk-extensions/autoconfigure/README.md
#   for opentelemetry options

java -javaagent:opentelemetry-javaagent.jar \
  -Dotel.metrics.exporter=none \
  -Dotel.service.name=middle-java \
  -Dotel.propagators=tracecontext,baggage,jaeger,b3,b3multi \
  -Dotel.traces.exporter=jaeger \
  -Dotel.exporter.jaeger.endpoint=http://localhost:14250/api/traces \
  -jar target/hacking-1.0-SNAPSHOT-jar-with-dependencies.jar \
  localhost


#IMAGE=docker.io/mfenwick100/hacking-java:latest
#
#docker run $IMAGE
