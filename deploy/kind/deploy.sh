#!/usr/bin/env bash

set -xv
set -euo pipefail

IMAGE=docker.io/mfenwick100/telemetry-hacking:latest
NS=th

kind create cluster --image=kindest/node:v1.23.4

kubectl create ns $NS
helm install my-jf jaeger --repo https://jaegertracing.github.io/helm-charts -n $NS

# prometheus?  grafana?

kind load docker-image $IMAGE

kubectl apply -f bottom.yaml -n $NS
kubectl apply -f middle.yaml -n $NS
kubectl apply -f top.yaml -n $NS
