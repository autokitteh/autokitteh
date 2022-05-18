#!/bin/bash

set -euo pipefail

ARCH="$(uname -m)"

export AKD_INIT_PATHS="${AKD_INIT_PATHS-"/work/examples/manifests"}"
export AKD_TEMPORAL_HOSTPORT="${AKD_TEMPORAL_HOSTPORT-host.docker.internal:7233}"

ENV_FILE="/tmp/autokitteh-$$.env"
trap "rm -f ${ENV_FILE}" 0

env | grep ^AKD_ > "${ENV_FILE}"

docker run \
  --env-file "${ENV_FILE}" \
  -v "${PWD}:/work" \
  -w "/work" \
  -p 20000:20000 \
  -p 20001:20001 \
  -it "autokitteh/autokitteh" \
  --enable defaults \
  "$@"
