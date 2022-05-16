#!/bin/bash

set -euo pipefail

mkdir -p /gen/openapi

find /gen/proto/src -mindepth 1 -type d | while read -r indir; do
  echo "grpcweb protoc: ${indir}"

  outdir="/gen/grpcweb/$(basename ${indir})"
  mkdir -p "${outdir}"

  /usr/bin/protoc \
    -I "/gen/proto/src" \
    -I "/proto" \
    "--js_out=import_style=commonjs:${outdir}" \
    "--grpc-web_out=import_style=commonjs,mode=grpcwebtext:${outdir}" \
    "${indir}"/*

done
