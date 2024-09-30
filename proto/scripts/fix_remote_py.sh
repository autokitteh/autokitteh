#!/bin/bash

set -euo pipefail

case "$OSTYPE" in
  darwin*|bsd*)
    sed_no_backup=( -i '' )
    ;; 
  *)
    sed_no_backup=( -i )
    ;;
esac

sed "${sed_no_backup[@]}" -e 's/from.*import/from \. import/' ../runtimes/pythonrt/runner/pb/autokitteh/remote/v1/remote_pb2_grpc.py
