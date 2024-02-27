#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

TESTS="${TESTS-*.txtar}"

AK="../../bin/ak"

for f in ${TESTS}; do
echo "- testing $f"
    ${AK} runtimes run --local --quiet --test "$f"
done
