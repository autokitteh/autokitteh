#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

# shellcheck disable=SC1091 
source common.sh

require npx "10.8.3" --version "\(.*\)"
