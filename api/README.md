Directories:
- **proto**: Source .proto files. [p2](https://github.com/wrouesnel/p2cli) is used to template some repeating definitions.
- **scripts**: Scripts (entrypoint is `scripts/gen.sh`) that generate **gen** using [autokitteh/protoc](https://hub.docker.com/r/autokitteh/protoc) docker image.
- **gen**: Generated stubs. `gen/src` are the post-p2 actual proto files.
