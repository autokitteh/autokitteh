#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

# shellcheck disable=SC1091 
source common.sh

require shellcheck "0.10.0" --version "version: \([0-9.]*\)"
