FROM python:3.11-slim AS base

RUN apt-get update && apt-get install --yes curl && rm -rf /var/lib/apt/lists/*

COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/

WORKDIR /runner

FROM base AS base_requirements
COPY runner/pyproject.toml .
RUN uv pip install .[all] --system

COPY py-sdk py-sdk
RUN cd py-sdk && uv pip install .[all] --system

FROM base_requirements AS custom_requirements

COPY workflow/user_requirements.txt .
RUN uv pip install -r user_requirements.txt --system


# https://stackoverflow.com/questions/78599865/how-to-install-missing-python-modules-on-distroless-image
FROM gcr.io/distroless/python3-debian12

USER 65532:65532

COPY --from=custom_requirements /usr/local/lib/python3.11/site-packages /usr/lib/python3.11/site-packages
COPY sitecustomize.py /usr/lib/python3.11/sitecustomize.py

COPY --chown=65532:65532 runner /runner
COPY --chown=65532:65532 workflow /workflow
