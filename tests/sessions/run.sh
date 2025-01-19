#!/bin/bash

set -euo pipefail

TESTS="${TESTS-*.txtar}"
PYTHON="${PYTHON-skip}"

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

    "${ak_filename}" --config "http.addr=:0" --config "runtimes.lazy_load_local_venv=true" --config "http.addr_filename=${addr_filename}" up -m test >& "${logfn}" &
    echo "waiting for autokitteh to be ready"

    ready=0
    while IFS= read -t 10 -r LL || [[ -n "$LL" ]]; do
        # TODO: maybe a more accurate way to parse.
        if [[ "${LL}" =~ "ready" ]]; then
            ready=1
            break
        fi
    done < <(tail -f "${logfn}")

    if [[ ${ready} -eq 0 ]]; then
        echo "autokitteh failed to start"
        cat "${logfn}"
        exit 1
    fi

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

for f in tests/sessions/${TESTS}; do
    name="$(basename "${f}")"

    echo "--- ${f} ---"

    case "${PYTHON}" in
        skip)
            if [[ ${name} == python_* ]]; then
                echo "skipping python test"
                continue
            fi
            ;;
        only)
            if [[ ${name} != python_* ]]; then
                echo "skipping non-python test"
                continue
            fi
            ;;
        *)
            ;;
    esac

    up "${name}"

    AK="${PWD}/${ak_filename} -C http.service_url=http://$(cat ${addr_filename})"

    ${AK} project create --name test

    ${AK} session test "${f}" --project test

    down 
done
