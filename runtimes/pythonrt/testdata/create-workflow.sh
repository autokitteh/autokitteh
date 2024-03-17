#!/bin/bash
# Create test workflow

set -e

ak manifest apply testdata/simple/autokitteh.yaml
tmp=$(mktemp)
ak project build py_simple --from testdata/simple/ | tee "${tmp}"
build_id=$(awk '{print $2}' "${tmp}")
env_id=$(ak envs list | grep default | awk -F\" '{print $2}')
ak deployment create --build-id "${build_id}" --activate --env "${env_id}"
