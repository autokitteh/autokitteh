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


replace autokitteh/module/v1/module_pb2.py "from buf." "from pb.buf."
replace autokitteh/user_code/v1/handler_svc_pb2.py "from autokitteh." "from pb.autokitteh." 
replace autokitteh/user_code/v1/handler_svc_pb2.pyi "from autokitteh." "from pb.autokitteh." 
replace autokitteh/user_code/v1/handler_svc_pb2_grpc.py "from autokitteh." "from pb.autokitteh." 
replace autokitteh/user_code/v1/runner_svc_pb2.py "from autokitteh." "from pb.autokitteh." 
replace autokitteh/user_code/v1/runner_svc_pb2.pyi "from autokitteh." "from pb.autokitteh." 
replace autokitteh/user_code/v1/runner_svc_pb2_grpc.py "from autokitteh." "from pb.autokitteh." 
replace autokitteh/values/v1/values_pb2.py "from autokitteh." "from pb.autokitteh." 
replace autokitteh/values/v1/values_pb2.py "from buf." "from pb.buf."
replace autokitteh/values/v1/values_pb2.pyi "from buf." "from pb.buf."
replace buf/validate/validate_pb2.py "from buf." "from pb.buf."
