#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

TESTS="${TESTS-*.txtar}"

AK="../../bin/ak"

run() {
    echo "- testing $f"
    ${AK} runtimes run --local --quiet --test "$f"
}

for f in ${TESTS}; do
    run "$f"
done
