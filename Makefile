export GOPRIVATE=github.com/autokitteh/*

GO=go
TAGS=

ARCH=$(shell uname -m)

ifndef GO_BUILD_OPTS
ifdef DEBUG
GO_BUILD_OPTS=-gcflags=all="-N -l"
else
GO_BUILD_OPTS=
endif
endif

OUTDIR?=bin
BUILD_OUTDIR=$(OUTDIR)
GENDIR?=gen

ifeq (, $(shell which gotestsum))
GOTEST=$(GO) test
else
GOTEST=gotestsum --
endif

define build
$(GO) build --tags "${TAGS}" -o $(BUILD_OUTDIR)/$@ $(GO_BUILD_OPTS) ./cmd/$@
endef

define test
$(GOTEST) -v $(GO_TEST_OPTS) -count=1 "$1"
endef

.PHONY: all
all: shellcheck api py bin lint test

.PHONY: clean
clean:
	rm -fR $(OUTDIR) $(GENDIR)
	mkdir $(OUTDIR) $(GENDIR)
	make -C tests clean

.PHONY: plugins
plugins: aktestplugind akpluginsd

.PHONY: bin
bin: akd plugins

.PHONY: build
build:
	$(GO) build $(GO_BUILD_OPTS) ./...
	make lint

.PHONY: debug
debug:
	GO_BUILD_OPTS='-gcflags=all="-N -l"' make bin

.PHONY: ak
ak:
	$(build)

.PHONY: dashboard
dashboard:
	cd web/dashboard && npm run build

.PHONY: akd
akd:
ifeq (,$(wildcard web/dashboard/build/index.html))
	@echo "*** WARNING: dashboard not built and will not be embedded"
	mkdir -p web/dashboard/build
	touch web/dashboard/build/.keep
endif
	$(build)

.PHONY: aksh
aksh:
	$(build)

.PHONY: akpluginsd
akpluginsd:
	$(eval BUILD_OUTDIR:=$(OUTDIR)/plugins)
	$(build)

.PHONY: aktestplugind
aktestplugind:
	$(eval BUILD_OUTDIR:=$(OUTDIR)/plugins)
	$(build)

.PHONY: d
d: akd

.PHONY: c
c: ak

.PHONY: sh
sh: aksh

.PHONY: $(GENDIR)/proto
$(GENDIR)/proto:
	rm -fR $(GENDIR)/proto
	./api/scripts/gen.sh

.PHONY: api
api: $(GENDIR)/proto

$(OUTDIR)/tools/golangci-lint:
	mkdir -p $(OUTDIR)/tools
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(OUTDIR)/tools" v1.44.2

$(OUTDIR)/tools/clitest:
	mkdir -p $(OUTDIR)/tools
	curl -o $@ -sOL https://raw.githubusercontent.com/aureliojargas/clitest/master/clitest
	chmod +x $@

.PHONY: tools
tools: $(OUTDIR)/tools/golangci-lint $(OUTDIR)/tools/clitest

.PHONY: test
test: test-unit test-cli

.PHONY: test-cli
test-cli: $(OUTDIR)/tools/clitest
	make -C tests test-cli

.PHONY: test-aksh
test-aksh:
	make -C tests test-aksh

.PHONY: test-unit
test-unit:
	$(GOTEST) -v --race --tags="unit" -count=1 $(or ${tests},${tests},./...)
	$(GOTEST) -v --tags="unit_norace" -count=1 $(or ${tests},${tests},./...)

.PHONY: lint
lint: $(OUTDIR)/tools/golangci-lint
	$(OUTDIR)/tools/golangci-lint run

.PHONY: shellcheck
shellcheck:
	docker run -v $(shell pwd):/src -w /src koalaman/shellcheck -a -- $(shell find . -name \*.sh)

.PHONY: docker-protoc
docker-protoc:
	make -C build/protoc

.PHONY: docker-autokitteh
docker-autokitteh:
	docker build -t autokitteh/autokitteh -f build/autokitteh/Dockerfile .

.PHONY: docker
docker: docker-autokitteh docker-protoc

.PHONY: py
py:
	make -C py

.PHONY: goreleaser
goreleaser:
	goreleaser release --snapshot --rm-dist
