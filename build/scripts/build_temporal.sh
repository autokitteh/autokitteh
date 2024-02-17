#!/bin/sh

BASEDIR="$(dirname "$0")"
DOCKERFILE_PATH=$BASEDIR/../temporal/Dockerfile
BUILD_CONTEXT=$BASEDIR/../../
docker build -t autokitteh/temporal -f "${DOCKERFILE_PATH} ${BUILD_CONTEXT}"
