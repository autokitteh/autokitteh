# shellcheck shell=bash

fail() {
    echo "error: $1"
    exit 1
}

require() {
    BIN="$1"
    REQ_VER="$2"
    VER_CMD="$3"
    VER_PATTERN="$4"

    if ! which "${BIN}" >& /dev/null; then
        fail "missing ${BIN} binary, please install golang ${REQ_VER} or later."
    fi

    CURR_VER="$($BIN "$VER_CMD" | sed -ne "s/${VER_PATTERN}/\1/p")"

    rc="$(./semver compare "$CURR_VER" "$REQ_VER")"

    if [[ $rc == "-1" ]]; then
        fail "${BIN} version ${CURR_VER} is too old, please install ${REQ_VER} or later."
    fi
}
