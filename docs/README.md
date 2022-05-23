# DANGER DANGER DANGER

DO NOT USE. NOT READY FOR GA.

# RUN USING DOCKER

```
# OPTIONAL: Build AutoKitteh docker image locally.
# If not done, latest image from dockerhub is used.
# Local build takes ~115s cold on an Apple M1 Pro 2021.
$ make docker

# Run temporal, might want to wait 30s to let it start up.
$ ./scripts/temporal/up.sh -d

# Preconfigured with some basic account:
$ ./scripts/akd-docker.sh --setup

# Clean slate:
$ ./scripts/akd-docker.sh

```

# RUN LOCAL BUILD, MINIMAL

```
# Requires go >= 1.18 installed locally on the machine.
$ make bin

# Run temporal, might want to wait 30s to let it start up.
$ ./scripts/temporal/up.sh -d

# Preconfigured:
$ ./scripts/akd.sh --setup

# Clean slate:
$ ./bin/akd
```

# API CHANGES

- Protobuffer defintions are in https://github.com/autokitteh/idl.
- SDK, which depends on the protobuffers, are in https://github.com/autokitteh/go-sdk.

# LOCAL BUILD REQUIREMENTS

## core

- go >= 1.18
- optional: docker (for shellcheck)
- optional: gotestsum
- optional: goreleaser (for testing releases)
- optional: tctl

## local protoc build

- jq
- docker
