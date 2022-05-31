#!/bin/bash

set -euo pipefail

if [[ ${DSN-} == grpc:* ]]; then
  ../../bin/akd --disable hello --enable defaults &
  pid=$!
  trap 'kill ${pid}' 0
  sleep 0.1
fi

../../bin/aksh "${@}"
