#!/bin/bash

set -euo pipefail

# This script does the same stuff that is being done by gen.sh, but from inside
# a docker build (build/autokitteh/Dockerfile) without a pre-made protoc image.

if [[ "${PWD}" -ne "/build" ]]; then
  echo "Must be run in /build in docker build context"
  exit 1
fi

chmod +x build/protoc/scripts/install.sh && ./build/protoc/scripts/install.sh

mkdir -p /gen/proto
mkdir -p /build/gen/proto/src
mkdir -p /build/gen/proto/stubs

cp -r /build/api/proto /proto/src
cp -r /build/gen/proto/src /gen/proto/src
cp -r /build/api/scripts /scripts

./api/scripts/_tmpl.sh
./api/scripts/_go.sh
./api/scripts/_openapi.sh

ln -s /gen/go/gitlab.com/softkitteh/autokitteh/gen/proto/stubs/go /build/gen/proto/stubs/go
ln -s /gen/openapi /build/gen/proto/openapi
