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

All generated code for proto is checked into the repository.
If any proto is changed, these must be ran:

```
# OPTIONAL: Build protoc docker image locally.
# If not done, latest image from dockerhub is used.
$ make protoc

# Always run.
$ make api
```

# LOCAL BUILD REQUIREMENTS

## core

- go >= 1.18
- optional: docker (for shellcheck)
- optional: gotestsum
- optional: goreleaser (for testing releases)

## local protoc build

- jq
- docker

## py

- python3
- pipenv

