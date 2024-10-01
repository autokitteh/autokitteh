# syntax=docker/dockerfile:1

################################################################################
# Create a stage for building the application.
ARG GO_VERSION=1.23
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
    export LDFLAGS="-X "${VERSION_PKG_PATH}.Version=$(cat .version || echo)" -X "${VERSION_PKG_PATH}.Time=${TIMESTAMP}" -X "${VERSION_PKG_PATH}.Commit=$(cat .commit || echo)""
    CGO_ENABLED=0 go build -o /bin/ak -ldflags="${LDFLAGS}" ./cmd/ak
EOF




################################################################################
# Create a new stage for running the application that contains the minimal
# runtime dependencies for the application.
FROM python:3.11-slim AS pydeps

WORKDIR /runner
COPY ./runtimes/pythonrt/runner/pyproject.toml pyproject.toml
RUN python -m pip install .[all]



FROM python:3.11-slim AS final

# Install any runtime dependencies 
# RUN --mount=type=cache,target=/var/cache/apk \
#     apk --update add \
#     ca-certificates \
#     tzdata \
#     && \
#     update-ca-certificates

# COPY ./scripts/gcc-on-arm.sh /tmp
# RUN /tmp/gcc-on-arm.sh
# RUN rm /tmp/gcc-on-arm.sh

# Create a non-privileged user 
# See https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#user
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/home/appuser" \
    --shell "/sbin/nologin" \
    --uid "${UID}" \
    appuser
USER appuser

COPY --chown=appuser:appuser --from=pydeps /usr/local/lib/python3.11/site-packages /usr/lib/python3.11/site-packages

# Create initial venv
# RUN python3 -m venv ~/.local/share/autokitteh/venv
# COPY ./runtimes/pythonrt/pyproject.toml pyproject.toml
# RUN ~/.local/share/autokitteh/venv/bin/python -m pip install .[all]


# Copy the executable from the "build" stage.
COPY --chown=appuser:appuser --from=build /bin/ak /bin/

ENV PYTHONPATH=/usr/lib/python3.11/site-packages
ENV AK_WORKER_PYTHON=/usr/local/bin/python
# Expose the port that the application listens on.
EXPOSE 9980

ENTRYPOINT [ "/bin/ak" ]
