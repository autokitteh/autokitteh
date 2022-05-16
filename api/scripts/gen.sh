#!/bin/bash

set -euo pipefail

PROTOC_IMAGE_NAME="protoc-$(uname -m)"

if [[ -z $(docker images -q "${PROTOC_IMAGE_NAME}") ]]; then
  make protoc
fi

rm -fR gen/proto
mkdir -p gen/proto/stubs

run() {
  docker run \
    --rm \
    -it \
    -v "${PWD}/api/proto:/proto/src:ro" \
    -v "${PWD}/gen/proto/src:/gen/proto/src" \
    -v "${PWD}/gen/proto/stubs/go:/gen/go/gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go" \
    -v "${PWD}/gen/proto/stubs/py:/gen/py" \
    -v "${PWD}/gen/proto/openapi:/gen/openapi" \
    -v "${PWD}/gen/proto/stubs/grpcweb:/gen/grpcweb" \
    -v "${PWD}/api/scripts:/scripts:ro" \
    "${PROTOC_IMAGE_NAME}" \
    "${@}"
}

run "${CMD-/scripts/_all.sh}"
