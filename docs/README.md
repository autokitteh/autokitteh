# DANGER DANGER DANGER

DO NOT USE. NOT READY FOR GA.

# RUN USING DOCKER

```
# Takes ~160s cold on an Apple M1 Pro 2021.
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
# Only required if first ever running on machine.
# (requires docker to run in experimental mode)
$ make protoc

# Only needed if never built before, or any api changes.
$ make api

# Requires go >= 1.18 installed locally on the machine.
$ make bin

# Run temporal, might want to wait 30s to let it start up.
$ ./scripts/temporal/up.sh -d

# Preconfigured:
$ ./scripts/akd.sh --setup

# Clean slate:
$ ./bin/akd
```

# RUN LOCAL BUILD, WITH WORKING PLUGINS + EVENT SOURCES

TODO

# LOCAL BUILD REQUIREMENTS

## core

- go >= 1.18
- optional: docker (for shellcheck)
- optional: gotestsum

## local protoc build

- jq
- docker

## py

- python3
- pipenv

