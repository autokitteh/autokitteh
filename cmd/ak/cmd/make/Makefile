UVX_DEV=uvx --with pytest --with autokitteh[all] $(if $(wildcard requirements.txt),--with-requirements requirements.txt)

.PHONY: all
all: lint typecheck format test

.PHONY: lint
lint:
	uvx ruff check

.PHONY: format
format:
	uvx ruff format --check

.PHONY: typecheck
typecheck:
	$(UVX_DEV) mypy --follow-untyped-imports  .

.PHONY: test
test:
	@if find . -name 'test_*.py' -o -name '*_test.py' | grep -q .; then \
		$(UVX_DEV) pytest; \
	else \
		echo "No Python test files found, skipping tests"; \
	fi

-include Makefile
