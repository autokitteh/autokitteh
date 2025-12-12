UVX=uvx --with autokitteh $(if $(wildcard requirements.txt),--with-requirements requirements.txt)

.PHONY: all
all: lint typecheck format

.PHONY: lint
lint:
	$(UVX) ruff check

.PHONY: format
format:
	$(UVX) ruff format --check

.PHONY: typecheck
typecheck:
	$(UVX) mypy --follow-untyped-imports  .

.PHONY: deploy
deploy:
	ak deploy --manifest autokitteh.yaml
