.PHONY: all
all: buf-format-check buf-lint buf-gen

.PHONY: buf-format-check
buf-format-check:
	buf format --diff --exit-code

# Abort when detecting violations during local builds
# (not a replacement for linting in the CI workflow).
.PHONY: buf-lint
buf-lint:
	buf lint

.PHONY: buf-build
buf-build:
	rm -rf gen
	buf build

.PHONY: buf-gen
buf-gen: buf-build
	buf generate --include-imports
	./scripts/fixpy.sh

.PHONY: clean
clean:
	rm -fR gen
	rm -fR dist

.PHONY: py-deps
py-deps:
	python3 -m pip install --upgrade poetry

.PHONY: py-dist
py-dist: py-deps
	rm -fR dist
	python3 -m poetry build

.PHONY: py-dist-publish
py-dist-publish-test: py-deps
	python3 -m poetry config repositories.test-pypi https://test.pypi.org/legacy/
	python3 -m poetry publish --repository test-pypi

.PHONY: py-grpc-tools
py-grpc-tools:
	python -m pip install 'grpcio-tools ~= 1.66'

.PHONY: remote-py
remote-py:
	python -m grpc_tools.protoc \
		-I remote \
		--python_out=../runtimes/pythonrt/runner \
		--pyi_out=../runtimes/pythonrt/runner \
		--grpc_python_out=../runtimes/pythonrt/runner \
		remote/remote.proto

.PHONY: remote-go
remote-go:
	mkdir -p ../runtimes/pythonrt/pb
	protoc \
		-I remote \
		--go_out=../runtimes/pythonrt/pb \
		--go_opt=paths=source_relative \
		--go-grpc_out=../runtimes/pythonrt/pb \
		--go-grpc_opt=paths=source_relative \
		remote/remote.proto

.PHONY: remote
remote: remote-py remote-go
