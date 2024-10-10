#!/bin/bash
set -euo pipefail

# Make directories in pb importable
find ../runtimes/pythonrt/runner/pb -type d -exec touch {}/__init__.py \;

# Change import path in grpc
sed 's/from autokitteh.remote.v1 import remote_pb2/from . import remote_pb2/' ../runtimes/pythonrt/runner/pb/autokitteh/remote/v1/remote_pb2_grpc.py > /tmp/fixed.py
mv /tmp/fixed.py ../runtimes/pythonrt/runner/pb/autokitteh/remote/v1/remote_pb2_grpc.py
