#!/bin/bash

set -euo pipefail

TESTS="${TESTS-*.txt}"

# if set, will not start its own ak instance.
AK_ADDR="${AK_ADDR-}"

if [[ ! -r go.mod ]]; then
    echo "must be run in repo root"
    exit 10
fi

ak_filename="bin/ak"

if [[ ! -x "${ak_filename}" ]]; then
    echo "${ak_filename} not present"
    exit 11
fi

logdir="tmp/testcli"

rm -fR "${logdir}"
mkdir -p "${logdir}"

addr_filename="/tmp/addr-$$"

up() {
    if [[ -n $AK_ADDR ]]; then
        echo "${AK_ADDR}" > "${addr_filename}"
        return
    fi

    logfn="${logdir}/$1.log"

    rm -f "${logfn}"

    echo "starting autokitteh"

    "${ak_filename}" --config "http.addr=:0" --config "http.addr_filename=${addr_filename}" up -m test >& "${logfn}" &

    echo "waiting for autokitteh to be ready"

    while IFS= read -r LL || [[ -n "$LL" ]]; do
        # TODO: maybe a more accurate way to parse.
        if [[ "${LL}" =~ "ready" ]]; then
            break
        fi
    done < <(tail -f "${logfn}")

    echo "autokitteh is ready"
}

down() {
    pkill -a -P "$$" >& /dev/null || true
}

ontrap() {
    rm -f ${addr_filename}

    pkill -a -P "$$" >& /dev/null || true
}

trap ontrap 0

export AK

for f in tests/cli/${TESTS}; do
    up "$(basename "${f}")"

    AK="${PWD}/${ak_filename} --url http://$(cat ${addr_filename})"

    echo "--- ${f} ---"
    ./scripts/clitest -1 "${f}"

    down 
done
