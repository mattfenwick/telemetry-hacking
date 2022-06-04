#!/usr/bin/env bash

set -xv
set -euo pipefail

IMAGE=docker.io/mfenwick100/hacking-java:latest

mvn clean compile assembly:single

java -jar target/hacking-1.0-SNAPSHOT-jar-with-dependencies.jar

docker build -t $IMAGE .

docker run $IMAGE
