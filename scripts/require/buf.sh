#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

# shellcheck disable=SC1091 
source common.sh

require buf "1.46.0" --version "\(.*\)"
