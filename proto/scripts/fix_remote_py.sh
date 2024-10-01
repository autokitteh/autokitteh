#!/bin/bash

# Make directories in pb importable
find ../runtimes/pythonrt/runner/pb -type d -exec touch {}/__init__.py \;

# Fix gRPC import
sed \
	-i 's/from autokitteh.remote.v1/from pb.autokitteh.remote.v1/' \
	../runtimes/pythonrt/runner/pb/autokitteh/remote/v1/remote_pb2_grpc.py

