.PHONY: all
all: test

.PHONY: clean
clean: clean-lang

.PHONY: test
test: test-lang

.PHONY: test-lang
test-lang:
	make -C lang test

.PHONY: clean-lang
clean-lang:
	make -C lang clean

