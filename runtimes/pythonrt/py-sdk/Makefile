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
	python -m pip install build twine
	python -m pip install .[all,dev]
	rm -rf build *.egg-info

build-docs:
	cd docs && ./gen-reqs.py
	$(MAKE) -C docs html

test:
	ruff check .
	ruff format --check
	python -m pytest tests

ci: install-deps test
