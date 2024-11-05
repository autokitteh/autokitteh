GO=go

# go install gotest.tools/gotestsum@latest
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

export AK_SYSTEST_USE_PROC_SVC=1
export PYTHONPATH=$(PWD)/runtimes/pythonrt/py-sdk

# 1. Detect unformatted Go files
# 2. Run shellcheck (shell scripts linter)
# 3. Download latest web platform
# 4. Rebuild protocol buffer stubs
# 5. Build the entire Go codebase
# 6. Run golangci-lint (Go linters)
# 7. Build AK binary with version and/or debug info
# 8. Run all automated tests (unit + integration)
all: gofmt-check shellcheck webplatform proto lint build bin/ak test

.PHONY: clean
clean:
	rm -rf $(OUTDIR)
	make -C web/webplatform clean

.PHONY: ak
ak: webplatform bin/ak

.PHONY: bin
bin: bin/ak

.PHONY: bin/ak
bin/ak: require/go
	$(GO) build --tags "${TAGS}" -o "$@" -ldflags="$(LDFLAGS)" $(GO_BUILD_OPTS) ./cmd/$(shell basename $@)

.PHONY: build
build: require/go
	mkdir -p $(OUTDIR)
	$(GO) build $(GO_BUILD_OPTS) ./...

.PHONY: debug
debug:
	DEBUG=1 make bin

.PHONY: gofmt-check
gofmt-check:
	test -z $(shell gofmt -l .) || exit 1

golangci_lint=$(shell which golangci-lint)

# https://golangci-lint.run/usage/install/#local-installation
# Keep the same version in "/.github/workflows/ci-go.yml"!
# See: https://github.com/golangci/golangci-lint
$(OUTDIR)/tools/golangci-lint:
	mkdir -p $(OUTDIR)/tools
ifeq ($(golangci_lint),)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(OUTDIR)/tools" v1.60.3
else
	ln -fs $(golangci_lint) $(OUTDIR)/tools/golangci-lint
endif

.PHONY: lint
lint: $(OUTDIR)/tools/golangci-lint
	$(OUTDIR)/tools/golangci-lint run

scripts=$(shell find . -name \*.sh -not -path "*/.venv/*")

.PHONY: shellcheck
shellcheck: require/shellcheck
	shellcheck -a $(scripts) -x

.PHONY: test
test: test-race test-runs test-sessions

.PHONY: test-dbgorm
test-dbgorm: require/go
	for dbtype in sqlite postgres; do \
		echo running for $$dbtype; \
	go test -v ./internal/backend/db/dbgorm -dbtype $$dbtype ; \
	done

# Skip a few Go unit-tests under "runtimes/pythonrt/" - either because they
# fails due to missing Python deps, or because they are very slow (20-30 sec).
# Note that this affects only Go CI in GitHub (which runs "make test-unit"),
# but not manual runs of "make" (which depend on "test-race"), or Python CI
# in GitHub (which uses "runtimes/pythonrt/Makefile").
.PHONY: test-unit
test-unit: require/go
	$(GOTEST) ./... -skip "(pyExports|pySvc|createVEnv)"

# Subset of "test-unit", for simplicity.
.PHONY: test-system
test-system: require/go
	$(GOTEST) ./tests/system

.PHONY: test-runs
test-runs:
	./tests/runs/run.sh

.PHONY: test-sessions
test-sessions:
	./tests/sessions/run.sh

.PHONY: test-cover
test-cover: require/go
	$(GOTEST) -covermode=atomic -coverprofile=tmp/cover.out ./...
	go tool cover -html=tmp/cover.out

.PHONY: test-race
test-race: require/go
	$(GOTEST) -race ./...

.PHONY: proto
proto: require/go require/buf
	make -C proto
	$(GO) build -v $(GO_BUILD_OPTS) ./proto/...
	$(GOTEST) ./proto

.PHONY: pythonrt
pythonrt: require/go
	make -C runtimes/pythonrt/

.PHONY: generate-migrations
generate-migrations: require/atlas
	@read -p "Enter migration name: " migration_name; \
	atlas migrate diff $$migration_name --env sqlite; \
	atlas migrate diff $$migration_name --env postgres

# Requires nodejs installed
.PHONY: tailwindcss
tailwindcss: require/npx
	npx --yes tailwindcss build -o web/static/tailwind.css

.PHONY: webplatform
webplatform: require/go
	make -C ./web/webplatform

.PHONY: require/go
require/go:
	@./scripts/require/go.sh

.PHONY: require/buf
require/buf:
	@./scripts/require/buf.sh

.PHONY: require/atlas
require/atlas:
	@./scripts/require/atlas.sh

.PHONY: require/npx
require/npx:
	@./scripts/require/npx.sh

.PHONY: require/shellcheck
require/shellcheck:
	@./scripts/require/shellcheck.sh
