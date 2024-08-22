# syntax=docker/dockerfile:1

################################################################################
# Create a stage for building the application.
ARG GO_VERSION=1.22
FROM golang:${GO_VERSION} AS build
WORKDIR /src

# Download dependencies as a separate step to take advantage of Docker's caching.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage bind mounts to go.sum and go.mod to avoid having to copy them into
# the container.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# Build the application.
# Leverage a cache mount to /go/pkg/mod/ to speed up subsequent builds.
# Leverage a bind mount to the current directory to avoid having to copy the
# source code into the container.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    <<EOF
    export VERSION_PKG_PATH="go.autokitteh.dev/autokitteh/internal/version"
    export TIMESTAMP="$(date -u "+%Y-%m-%dT%H:%MZ")"
    export LDFLAGS="-s -w -X "${VERSION_PKG_PATH}.Version=$(cat .version || echo)" -X "${VERSION_PKG_PATH}.Time=${TIMESTAMP}" -X "${VERSION_PKG_PATH}.Commit=$(cat .commit || echo)""
    CGO_ENABLED=0 go build -o /bin/ak -ldflags="${LDFLAGS}" ./cmd/ak
EOF

FROM python:3.12 AS python_deps

WORKDIR /app

# RUN --mount=type=cache,target=/var/cache/apk \
#     apt-get \
#     g++ libstdc++ \
#     ca-certificates \
#     tzdata \
#     && \
#     update-ca-certificates

COPY ./scripts/gcc-on-arm.sh /tmp
RUN /tmp/gcc-on-arm.sh
RUN rm /tmp/gcc-on-arm.sh


RUN python3 -m venv env
COPY ./runtimes/pythonrt/requirements.txt requirements.txt
ENV GRPC_PYTHON_BUILD_SYSTEM_OPENSSL=1
ENV GRPC_PYTHON_BUILD_SYSTEM_ZLIB=1
RUN /app/env/bin/python3 -m pip install --no-cache-dir -r requirements.txt

################################################################################
# Create a new stage for running the application that contains the minimal
# runtime dependencies for the application.

FROM python:3.12-alpine3.20 AS final


# Create a non-privileged user 
# See https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#user
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/usr/src/autokitteh" \
    --shell "/sbin/nologin" \
    --uid "${UID}" \
    appuser

WORKDIR /usr/src/autokitteh
# Copy the executable from the "build" stage.
COPY --chown=appuser:appuser --from=python_deps /app/env /usr/src/autokitteh/.local/share/autokitteh/venv
COPY --from=build /bin/ak /bin/

# Expose the port that the application listens on.
EXPOSE 9980

USER appuser

ENTRYPOINT [ "/bin/ak" ]
