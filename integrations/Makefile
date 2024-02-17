# TODO: https://linear.app/autokitteh/issue/ENG-87/standardize-makefiles-across-repos

GO=go

OUTDIR?=bin

ifndef GO_BUILD_OPTS
ifdef DEBUG
GO_BUILD_OPTS=-gcflags=all="-N -l"
else
GO_BUILD_OPTS=
endif
endif

golangci_lint=$(shell which golangci-lint)

# go install gotest.tools/gotestsum@latest
ifeq ($(shell which gotestsum),)
GOTEST=$(GO) test
else
GOTEST=gotestsum -f testname --
endif

.PHONY: all
all: build check-gofmt lint test

.PHONY: clean
clean:
	rm -rf $(OUTDIR)

.PHONY: build
build:
	mkdir -p $(OUTDIR)
	$(GO) build -v $(GO_BUILD_OPTS) ./...

.PHONY: check-gofmt
check-gofmt:
	test -z $(shell gofmt -l .) || exit 1

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: test-cover
test-cover:
	$(GOTEST) -covermode=atomic -coverprofile=tmp/cover.out ./...
	go tool cover -html=tmp/cover.out

.PHONY: test-race
test-race:
	$(GOTEST) -race -v ./...

.PHONY: lint
lint: $(OUTDIR)/tools/golangci-lint
	$(OUTDIR)/tools/golangci-lint run

# https://golangci-lint.run/usage/install/#local-installation
# Keep the same version in the YAML files under ".github/workflows"!
# See: https://github.com/golangci/golangci-lint
$(OUTDIR)/tools/golangci-lint:
	mkdir -p $(OUTDIR)/tools
ifeq ($(golangci_lint),)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(OUTDIR)/tools" v1.55.2
else
	ln -fs $(golangci_lint) $(OUTDIR)/tools/golangci-lint
endif
