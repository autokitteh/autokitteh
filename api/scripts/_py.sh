#!/bin/bash

set -euo pipefail

mkdir -p /gen/py

gen() {
  python3 \
    -m grpc_tools.protoc \
    -I/gen/proto/src \
    -I/proto \
    --python_out=/gen/py \
    --grpc_python_out=/gen/py \
    --mypy_out=/gen/py \
    "${@}"
}

echo "py protoc: /proto/validate/validate"
gen /proto/validate/validate.proto

echo "py google.api.annotations: /proto/google/api/annotations"
gen /proto/google/api/annotations.proto

echo "py google.api.http: /proto/google/api/http"
gen /proto/google/api/http.proto

mv /gen/py/google/api /gen/py/googleapi && rm -fR /gen/py/google

find /gen/proto/src -mindepth 1 -type d | while read -r indir; do
  echo "py protoc: ${indir}"

  gen "${indir}"/*
done
