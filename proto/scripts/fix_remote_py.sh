#!/bin/bash
set -euo pipefail

out_dir="../runtimes/pythonrt/runner/pb"

# Make directories in pb importable
find ${out_dir} -type d -exec touch {}/__init__.py \;

# Change import path in grpc

replace() {
    sed "s/$2/$3/" "${out_dir}/$1" > "${out_dir}/$1.new"
    mv -f "${out_dir}/$1.new" "${out_dir}/$1"
}


replace autokitteh/remote/v1/remote_pb2_grpc.py "from autokitteh.remote.v1 import remote_pb2" "from . import remote_pb2"

replace autokitteh/values/v1/values_pb2.py "from autokitteh." "from pb.autokitteh." 
replace autokitteh/remote/v1/remote_pb2.py "from autokitteh." "from pb.autokitteh." 
replace buf/validate/validate_pb2.py "from buf." "from pb.buf."
replace autokitteh/module/v1/module_pb2.py "from buf." "from pb.buf."
replace autokitteh/values/v1/values_pb2.py "from buf." "from pb.buf."
