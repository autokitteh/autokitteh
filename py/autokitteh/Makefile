.PHONY: all
all: clean pipenv proto mypy test dist

.PHONY: clean
clean:
	pipenv clean
	rm -fR build dist src/autokitteh/proto src/autokitteh/*.egg_info

.PHONY: pipenv
pipenv:
	pipenv install

.PHONY: proto
proto:
	rm -fR src/autokitteh/proto && mkdir -p src/autokitteh/proto
	# bring in only specific protos.
	for p in $(shell cat PROTOS | grep -v "^#"); do cp -r "../../gen/proto/stubs/py/$${p}" "src/autokitteh/proto/$${p}"; done
	# fix internal imports to be relative imports.
	find . -name \*_pb2\*.py -exec sed -i '' 's/^from \([^.]*\) import \(.*\)_pb2/from ..\1 import \2_pb2/' {} \;
	find . -name \*_pb2\*.py -exec sed -i '' 's/^from google.api import \(.*\)_pb2/from ..googleapi import \1_pb2/' {} \;
	# add __init__.py so setup.py could find these packages.
	find src/autokitteh/proto -type d -mindepth 1 -exec touch {}/__init__.py \;

.PHONY: mypy
mypy:
	pipenv run mypy --strict src/autokitteh

.PHONY: test
test:
	pipenv run pytest -s .

.PHONY: dist
dist: proto
	rm -fR dist
	pipenv run python3 -m build

.PHONY: upload-test
upload-test:
	pipenv run python3 -m twine upload --repository testpypi dist/*

