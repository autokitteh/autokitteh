#!/bin/bash

set -euo pipefail

exec docker-compose \
  -f scripts/temporal/docker-compose-postgres.yml \
  up "$@"
