#!/bin/bash

set -euo pipefail

ARCH="$(uname -m)"

set -x

exec docker run \
  --env AKD_INIT_PATHS="${AKD_INIT_PATH-"/work/examples/manifests/default"}" \
  --env AKD_TEMPORAL_HOSTPORT="host.docker.internal:7233" \
  -v "${PWD}:/work" \
  -p 50000:50000 \
  -p 50001:50001 \
  -it "autokitteh-${ARCH}" \
  --enable defaults \
  "$@"
