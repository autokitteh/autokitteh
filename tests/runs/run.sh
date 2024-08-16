#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

TESTS="${TESTS-*.txtar}"
PYTHON="${PYTHON-skip}"

AK="../../bin/ak"

for f in ${TESTS}; do
    echo "- testing $f"

    case "${PYTHON}" in
        skip)
            if [[ ${f} == python_* ]]; then
                echo "skipping python test"
                continue
            fi
            ;;
        only)
            if [[ ${f} != python_* ]]; then
                echo "skipping non-python test"
                continue
            fi
            ;;
        *)
            ;;
    esac


    ${AK} runtime test --local --quiet "$f" && echo PASS || echo FAIL
done
