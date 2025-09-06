GO=go

ifeq ($(shell which gotestsum),)
GOTEST=$(GO) test
else
GOTEST=gotestsum -f testname --
endif

TAGS?=
OUTDIR?=bin

# https://go.dev/doc/gdb
# https://github.com/golang/vscode-go/blob/master/docs/debugging.md
ifndef GO_BUILD_OPTS
ifdef DEBUG
GO_BUILD_OPTS=-gcflags=all="-N -l"
else
GO_BUILD_OPTS=-trimpath
endif
endif

ARCH=$(shell uname -m)

ifeq ($(COMMIT),)
COMMIT=$(shell git rev-parse HEAD)$(shell git diff --quiet || echo '_dirty')
endif

ifeq ($(TIMESTAMP),)
TIMESTAMP=$(shell date -u "+%Y-%m-%dT%H:%MZ")
endif

ifeq ($(VERSION),)
VERSION=dev
endif

VERSION_PKG_PATH="go.autokitteh.dev/autokitteh/internal/version"
LDFLAGS+=-X '${VERSION_PKG_PATH}.Version=${VERSION}' -X '${VERSION_PKG_PATH}.Time=${TIMESTAMP}' -X '${VERSION_PKG_PATH}.Commit=${COMMIT}' -X '${VERSION_PKG_PATH}.User=$(shell whoami)' -X '${VERSION_PKG_PATH}.Host=$(shell hostname)'

export PYTHONPATH=$(PWD)/runtimes/pythonrt/py-sdk

.PHONY: ak
ak: webplatform bin/ak

# 1. Detect unformatted Go files
# 2. Run shellcheck (shell scripts linter)
# 3. Download latest web platform
# 4. Rebuild protocol buffer stubs
# 5. Run golangci-lint (Go linters)
# 6. Build AK binary with version and/or debug info
# 7. Run all automated tests (unit + integration)
all: gofmt-check shellcheck webplatform proto lint bin/ak test

.PHONY: clean
clean:
	rm -rf $(OUTDIR)
	make -C web/webplatform clean

.PHONY: bin
bin: bin/ak

.PHONY: bin/ak
bin/ak:
	$(GO) build --tags "${TAGS}" -o "$@" -ldflags="$(LDFLAGS)" $(GO_BUILD_OPTS) ./cmd/$(shell basename $@)

.PHONY: debug
debug:
	DEBUG=1 make bin

.PHONY: gofmt-check
gofmt-check:
	test -z $(shell gofmt -l .) || exit 1

.PHONY: lint
lint:
	go tool golangci-lint run

scripts=$(shell find . -name \*.sh -not -path "*/.venv/*")

.PHONY: shellcheck
shellcheck:
ifneq ($(scripts),)
	docker run --rm -v $(shell pwd):/src -w /src koalaman/shellcheck:stable -a $(scripts) -x
endif

.PHONY: test
test: test-race test-db test-opa test-starkark test-sessions

# Run only Go unit-tests, without checking for race conditions,
# and without running long-running Python runtime and system tests.
.PHONY: test-unit
test-unit:
	$(GOTEST) $(go list ./... | grep -v -E "autokitteh/tests|runtimes/python")

# Run all Go tests (including Python runtime and system tests),
# and check for race conditions while running each of them.
.PHONY: test-race
test-race:
	$(GOTEST) -timeout 0 -race ./...

# Generate a coverage report for all Go tests
# (including Python runtime and system tests).
.PHONY: test-cover
test-cover:
	$(GOTEST) -covermode=atomic -coverprofile=tmp/cover.out -timeout 0 ./...
	go tool cover -html=tmp/cover.out

# Long-running subset of "test-unit", for simplicity.
.PHONY: test-system
test-system: bin/ak
	AK_SYSTEST_USE_PROC_SVC=1 $(GOTEST) -timeout 0 ./tests/system

.PHONY: test-db
test-db:
	for dbtype in sqlite postgres; do \
		echo running for $$dbtype; \
	$(GOTEST) ./internal/backend/db/... -dbtype $$dbtype ; \
	done

.PHONY: test-opa
test-opa:
	@if which opa > /dev/null; then \
		opa test configs/opa_bundles -v; \
	else \
		echo "opa not found, skipping OPA tests"; \
	fi

.PHONY: test-starlark
test-starlark: bin/ak
	$(GOTEST) ./tests/starlark/...

.PHONY: test-sessions
test-sessions: bin/ak
	$(GOTEST) ./tests/sessions/...
	$(GOTEST) ./tests/sessions/... --durable

.PHONY: proto
proto:
	make -C proto
	$(GO) build -v $(GO_BUILD_OPTS) ./proto/...
	$(GOTEST) ./proto

.PHONY: pythonrt
pythonrt:
	make -C runtimes/pythonrt/

.PHONY: generate-migrations
generate-migrations:
	@read -p "Enter migration name: " migration_name; \
	atlas migrate diff $$migration_name --env sqlite; \
	atlas migrate diff $$migration_name --env postgres; \
	atlas migrate diff $$migration_name --env postgres_enterprise;


# Requires nodejs installed
.PHONY: tailwindcss
tailwindcss:
	npx --yes tailwindcss build -o web/static/tailwind.css

.PHONY: webplatform
webplatform:
	make -C ./web/webplatform
