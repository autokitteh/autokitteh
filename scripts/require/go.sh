#!/bin/bash

set -euo pipefail

REQ_VER="$(grep ^go go.mod | cut -d\  -f 2)"

cd "$(dirname "$0")"

# shellcheck disable=SC1091 
source common.sh

require go "${REQ_VER}" version "go version go\(.*\) .*$"
