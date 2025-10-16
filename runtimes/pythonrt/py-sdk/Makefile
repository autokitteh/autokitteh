default:
	$(error please pick a target)

.PHONY: build
build:
	rm -rf build dist
	python -m build

# You need to set PYPI_TOKEN in your environment, see 1password to the token
upload: build
	twine upload \
		--username __token__ \
		--password $(PYPI_TOKEN) \
		dist/autokitteh-*

install-deps:
	uv pip install build twine
	uv pip install .[all,dev]
	rm -rf build *.egg-info

build-docs:
	cd docs && ./gen-reqs.py
	cd docs && ./gen-index.py
	$(MAKE) -C docs html

test:
	uv run ruff check .
	uv run ruff format --check
	uv run python -m pytest tests

venv:
	uv venv

ci: venv install-deps test
