#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

VERSION="$(cut -d\  -f 1 < VERSION)"
EXPECTED_SHA="$(cut -d\  -f 2 < VERSION)"
FILENAME="autokitteh-web-v${VERSION}.zip"

echo "want version ${VERSION}."

checksum() {
    echo "verifying checksum..."

    sha=$(shasum -a 256 "$1" | cut -d ' ' -f 1)

    if [[ "${sha}" != "${EXPECTED_SHA}" ]]; then
        echo "checksum mismatch: ${sha} != ${EXPECTED_SHA}"
        exit 1
    fi
}

if [[ -r ${FILENAME} ]]; then
    echo "required version already downloaded"
    checksum "${FILENAME}"
    echo "checksum verified"
    exit 0
fi

echo "not downloaded yet."

echo "fetching ${FILENAME}..."

trap 'rm -f ${FILENAME}_' 0

curl -sL "https://github.com/autokitteh/web-platform/releases/download/v${VERSION}/${FILENAME}" > "${FILENAME}_"

checksum "${FILENAME}_"

echo "removing previous versions..."
rm -f "autokitteh-web-v*.zip"

echo "finalizing version..."
mv "${FILENAME}_" "${FILENAME}"

echo "testing..."

go test -v ./...

echo "done!"
