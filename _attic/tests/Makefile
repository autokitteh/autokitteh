.PHONY: all
all: test

.PHONY: clean
clean: clean-cli

.PHONY: test
test: test-cli test-aksh

.PHONY: test-aksh
test-aksh:
	make -C aksh test

.PHONY: test-cli
test-cli:
	make -C cli test

.PHONY: clean-cli
clean-cli:
	make -C cli clean
