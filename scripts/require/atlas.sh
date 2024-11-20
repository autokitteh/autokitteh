#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

# shellcheck disable=SC1091 
source common.sh

require atlas "0.28.2" version "atlas version v\([0-9.]*\)-.*$"
