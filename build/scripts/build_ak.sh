#!/bin/sh

BASEDIR="$(dirname "$0")"
DOCKERFILE_PATH=$BASEDIR/../ak/Dockerfile
BUILD_CONTEXT=$BASEDIR/../../
docker build -t autokitteh/server -f "${DOCKERFILE_PATH} ${BUILD_CONTEXT}"
