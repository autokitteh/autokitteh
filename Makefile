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
GO_BUILD_OPTS=
endif
endif

ARCH=$(shell uname -m)

ifeq ($(COMMIT),)
COMMIT=$(shell git rev-parse HEAD)
endif

ifeq ($(TIMESTAMP),)
TIMESTAMP=$(shell date -u "+%Y-%m-%dT%H:%MZ")
endif

ifeq ($(VERSION),)
VERSION=dev
endif

VERSION_PKG_PATH="go.autokitteh.dev/autokitteh/internal/version"
LDFLAGS+=-X '${VERSION_PKG_PATH}.Version=${VERSION}' -X '${VERSION_PKG_PATH}.Time=${TIMESTAMP}' -X '${VERSION_PKG_PATH}.Commit=${COMMIT}$(shell git diff --quiet || echo '(dirty)')' -X '${VERSION_PKG_PATH}.User=$(shell whoami)' -X '${VERSION_PKG_PATH}.Host=$(shell hostname)'

export AK_SYSTEST_USE_PROC_SVC=1

# 1. Detect unformatted Go files
# 2. Run golangci-lint (Go linters)
# 3. Run shellcheck (shell scripts linter)
# 4. Rebuild protocol buffer stubs
# 5. Build the entire Go codebase
# 6. Build AK binary with version and/or debug info
# 7. Run all automated tests (unit + integration)
all: gofmt-check lint shellcheck proto build bin/ak test

.PHONY: clean
clean:
	rm -rf $(OUTDIR)

.PHONY: ak
ak: bin/ak

.PHONY: bin
bin: bin/ak

.PHONY: bin/ak
bin/ak:
	$(GO) build --tags "${TAGS}" -o "$@" -ldflags="$(LDFLAGS)" $(GO_BUILD_OPTS) ./cmd/$(shell basename $@)

.PHONY: build
build:
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
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(OUTDIR)/tools" v1.56.2
else
	ln -fs $(golangci_lint) $(OUTDIR)/tools/golangci-lint
endif

.PHONY: lint
lint: $(OUTDIR)/tools/golangci-lint
	$(OUTDIR)/tools/golangci-lint run

scripts=$(shell find . -name \*.sh)

.PHONY: shellcheck
shellcheck:
ifneq ($(scripts),)
	docker run --rm -v $(shell pwd):/src -w /src koalaman/shellcheck:stable -a $(scripts) -x
endif

.PHONY: test
test: test-race test-cli test-runs test-sessions

# TODO(ENG-427): Fix E2E test's data race.
# TODO(ENG-447): Fix HTTP trigger flakiness.
.PHONY: test-unit
test-unit:
	$(GOTEST) ./... -skip /workflows/builtin_funcs

# Subset of "test-unit", for simplicity.
.PHONY: test-system
test-system:
	$(GOTEST) ./tests/system

.PHONY: test-runs
test-runs:
	./tests/runs/run.sh

.PHONY: test-sessions
test-sessions:
	./tests/sessions/run.sh

.PHONY: test-cover
test-cover:
	$(GOTEST) -covermode=atomic -coverprofile=tmp/cover.out ./...
	go tool cover -html=tmp/cover.out

# TODO(ENG-427): Fix E2E test's data race.
# TODO(ENG-447): Fix HTTP trigger flakiness.
.PHONY: test-race
test-race:
	$(GOTEST) -race ./... -skip /workflows/builtin_funcs

.PHONY: test-cli
# We don't want test-cli to explicitly depend on bin since
# we might run test-cli multiple times to debug with the
# same build.
test-cli:
	./tests/cli/run.sh

.PHONY: proto
proto:
	make -C proto
	$(GO) build -v $(GO_BUILD_OPTS) ./proto/...
	$(GOTEST) ./proto
