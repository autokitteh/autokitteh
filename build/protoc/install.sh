#!/bin/bash

# run only inside docker.

set -euo pipefail

ARCH="$(uname -m)"
case "${ARCH}" in
  "aarch64" | "arm64")
    PROTOC_ARCH="aarch_64";;
  *)
    PROTOC_ARCH="${ARCH}";;
esac

PROTOC_VER=3.17.3
GRPC_VER=1.1.0
PROTOC_GEN_GO_VER=1.27.1
PROTOC_GEN_VALIDATE_VER=0.6.1
GRPC_GATEWAY_VER=2.5.0
GNOSTIC_VER=0.6.8
GRPC_WEB_VER=1.3.1
LIBPROTOCDEV_VER="3.12.4-1"

apt-get update -qq

apt-get -y install python3
apt-get -y install python3-setuptools
apt-get -y install python3-pip

pip3 install grpcio-tools mypy-protobuf

apt-get install unzip -yq

apt-get -y install "libprotoc-dev=${LIBPROTOCDEV_VER}" # needed for protoc-web

mkdir /proto

mkdir /go/src/protoc

cd /go/src/protoc

go mod init github.com/autokitteh/autokitteh/protoc

go_install() {
  go install "${1}@${2}"
}

# gomplate
go_install github.com/wrouesnel/p2cli/cmd/p2 r11

# protoc
wget -O/tmp/protoc.zip "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VER}/protoc-${PROTOC_VER}-linux-${PROTOC_ARCH}.zip"
unzip -o /tmp/protoc.zip -d /usr
rm -f /tmp/protoc.zip

# protoc-gen-validate
go_install "github.com/envoyproxy/protoc-gen-validate" "v${PROTOC_GEN_VALIDATE_VER}"
mkdir /proto/validate
wget -O /proto/validate/validate.proto "https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/v${PROTOC_GEN_VALIDATE_VER}/validate/validate.proto"

# protoc-gen-openapi
go_install github.com/google/gnostic/cmd/protoc-gen-openapi "v${GNOSTIC_VER}"

# grpc
go_install "google.golang.org/grpc/cmd/protoc-gen-go-grpc" "v${GRPC_VER}"

# protoc-gen-go
go_install "google.golang.org/protobuf/cmd/protoc-gen-go" "v${PROTOC_GEN_GO_VER}"

# grpc-gateway
go_install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway" "v${GRPC_GATEWAY_VER}"
go_install "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2" "v${GRPC_GATEWAY_VER}"

# grpc-web
# (need to build from scratch since no official arm64 version, see https://github.com/grpc/grpc-web/issues/1159)
# TODO: move to official arm64 version.
cd /tmp
wget "https://github.com/grpc/grpc-web/archive/refs/tags/${GRPC_WEB_VER}.tar.gz"
tar xvf "${GRPC_WEB_VER}.tar.gz"
cd "grpc-web-${GRPC_WEB_VER}/javascript/net/grpc/web/generator"
make
mv protoc-gen-grpc-web /usr/local/bin

# google apis
# see https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/#using-protoc
cd /tmp
wget https://github.com/googleapis/googleapis/archive/refs/heads/master.zip
unzip master.zip
mv googleapis-master/google /proto/google

rm -fR /tmp/*
