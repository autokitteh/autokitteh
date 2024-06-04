default:
	$(error please pick a target)

build:
	rm -rf build
	python -m build

# You need to set PYPI_TOKEN in your environment, see 1password to the token
upload: build
	twine upload \
		--username __token__ \
		--password $(PYPI_TOKEN) \
		dist/autokitteh-*

install-deps:
	python -m pip install build twine
