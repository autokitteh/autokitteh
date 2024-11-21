#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

VERSION="$(awk '{ print $1 }' < VERSION)"
EXPECTED_SHA="$(awk '{ print $2 }' < VERSION)"
FILENAME="autokitteh-web-v${VERSION}.zip"

echo "want version ${VERSION}."

checksum() {
    if [[ -z "${EXPECTED_SHA}" ]]; then
        echo "no checksum provided"
        return
    fi

    echo "verifying checksum..."

    sha=$(shasum -a 256 "$1" | awk '{ print $1 }')

    if [[ "${sha}" != "${EXPECTED_SHA}" ]]; then
        echo "checksum mismatch: ${sha} != ${EXPECTED_SHA}"
        exit 1
    fi

    echo "checksum verified"
}

if [[ -r ${FILENAME} ]]; then
    echo "required version already downloaded"
    checksum "${FILENAME}"
    exit 0
fi

echo "not downloaded yet."

echo "fetching ${FILENAME}..."

trap 'rm -f ${FILENAME}_' 0

curl -sL "https://github.com/autokitteh/web-platform/releases/download/v${VERSION}/${FILENAME}" > "${FILENAME}_"

checksum "${FILENAME}_"

echo "removing previous versions..."
rm -f autokitteh-web-v*.zip # DO NOT QUOTE. This is a glob.

echo "finalizing version..."
mv "${FILENAME}_" "${FILENAME}"

echo "testing..."

go test -v ./...

echo "done!"
