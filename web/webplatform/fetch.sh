#!/bin/bash

set -euo pipefail

cd "$(dirname "$0")"

VERSION="$(cat VERSION)"
FILENAME="autokitteh-web-v${VERSION}.zip"

echo "want version ${VERSION}"

if [[ -r ${FILENAME} ]]; then
    echo "up to date"
    exit 0
fi

echo "not downloaded yet."

PREV="$(ls -1 autokitteh-web-v*.zip > /dev/null 2>&1 || true)"

echo "fetching ${FILENAME}..."

curl -sL "https://github.com/autokitteh/web-platform/releases/download/v${VERSION}/${FILENAME}" > "${FILENAME}"

if [[ -z ${PREV} ]]; then
    echo "removing previous versions..."
    rm -f "${PREV}"
fi

echo "testing..."

go test -v ./...

echo "done!"
