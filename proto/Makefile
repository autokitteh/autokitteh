.PHONY: all
all: buf-format-check buf-lint buf-gen remote

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
	python3 -m pip install --upgrade poetry 'grpcio-tools ~= 1.66'

.PHONY: py-dist
py-dist: py-deps
	rm -fR dist
	python3 -m poetry build

.PHONY: py-dist-publish
py-dist-publish-test: py-deps
	python3 -m poetry config repositories.test-pypi https://test.pypi.org/legacy/
	python3 -m poetry publish --repository test-pypi

remote_pb_dir = autokitteh/remote/v1
remote_pb = $(remote_pb_dir)/remote.proto
remote_py_out = ../runtimes/pythonrt/runner
remote_go_out = gen/go/autokitteh/remote/v1

.PHONY: remote-py
remote-py:
	python -m grpc_tools.protoc \
		-I $(remote_pb_dir) \
		--python_out=$(remote_py_out)\
		--pyi_out=$(remote_py_out) \
		--grpc_python_out=$(remote_py_out) \
		$(remote_pb)

.PHONY: remote-go
remote-go:
	mkdir -p $(remote_go_out)
	protoc \
		-I $(remote_pb_dir) \
		--go_out=$(remote_go_out) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(remote_go_out) \
		--go-grpc_opt=paths=source_relative \
		$(remote_pb)

.PHONY: remote
remote: remote-py remote-go
