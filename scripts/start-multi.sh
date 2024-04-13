#!/bin/bash

# Requirements:
# - nginx installed
# - temporal server start-dev

set -euo pipefail

WD="${WD-tmp/multi}"

rm -fR "${WD}" && mkdir -p "${WD}"

./bin/ak server setup -m dev

akup() {
    port="$1"
    ./bin/ak up -m dev \
        -c "temporalclient.start_dev_server_if_not_up=false" \
        -c "http.addr=0.0.0.0:${port}" \
        -c "logger.zap.outputPaths=${WD}/${port}.log" \
        >& "${WD}/${port}.out" &
    echo $!
}

p1=$(akup 9981)
p2=$(akup 9982)
p3=$(akup 9983)

trap 'kill $p1 $p2 $p3' 0

nginx -c "$(pwd)/scripts/nginx.conf"
