#!/bin/bash

set -euo pipefail

rm -fR gen/proto
mkdir -p gen/proto/src

for stage in tmpl openapi grpcweb go py; do
  "$(dirname "$0")/_${stage}.sh"
done
