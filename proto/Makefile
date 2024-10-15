runner_pb_dir="../../runtimes/pythonrt/runner/pb"

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
	rm -fR $(runner_pb_dir)
	buf build

.PHONY: buf-gen
buf-gen: buf-build
	buf generate --include-imports
	./scripts/fixpy.sh
	buf generate --template buf.remote.gen.yaml --include-imports
	./scripts/fix_remote_py.sh

	

.PHONY: clean
clean:
	rm -fR gen
	rm -fR dist
	rm -fR $(runner_pb_dir)
.PHONY: py-dist

py-dist: py-deps
	rm -fR dist
	python3 -m poetry build

.PHONY: py-dist-publish
py-dist-publish-test: py-deps
	python3 -m poetry config repositories.test-pypi https://test.pypi.org/legacy/
	python3 -m poetry publish --repository test-pypi
