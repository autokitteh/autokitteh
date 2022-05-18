#!/bin/bash

set -euo pipefail

PROTOC_IMAGE_NAME="autokitteh/protoc"

rm -fR api/gen
mkdir -p api/gen

run() {
  docker run \
    --rm \
    -it \
    -v "${PWD}/api/proto:/proto/src:ro" \
    -v "${PWD}/api/gen/src:/gen/proto/src" \
    -v "${PWD}/api/gen/stubs/go:/gen/go/github.com/autokitteh/autokitteh/api/gen/stubs/go" \
    -v "${PWD}/api/gen/stubs/py:/gen/py" \
    -v "${PWD}/api/gen/openapi:/gen/openapi" \
    -v "${PWD}/api/gen/stubs/grpcweb:/gen/grpcweb" \
    -v "${PWD}/api/scripts:/scripts:ro" \
    "${PROTOC_IMAGE_NAME}" \
    "${@}"
}

run "${CMD-/scripts/_all.sh}"
