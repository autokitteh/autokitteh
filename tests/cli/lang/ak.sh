#!/bin/bash

set -euo pipefail

cli_flags=
if [[ -n ${test_grpc-} ]]; then
  AKD_CATALOG_PERMISSIVE=1 AKD_LOG_LEVEL=error ../../../bin/akd --enable testidgen --disable hello &
  pid=$!
  trap 'kill ${pid}' 0
  cli_flags="-a 127.0.0.1:50001"

  sleep 0.1
fi

# shellcheck disable=SC2086
../../../bin/ak ${cli_flags} "${@}"
