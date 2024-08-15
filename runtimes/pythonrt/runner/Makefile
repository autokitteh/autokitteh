default:
	$(error please pick a target)

test:
	ruff check .
	ruff format --check
	python -m pytest

install-deps:
	python -m pip install .[all]
	rm -rf build *.egg-info

ci: install-deps test

call-start:
	grpcurl --plaintext -d @ localhost:9293 Runner.Start < tests/start.json

call-exports:
	grpcurl --plaintext -d @ localhost:9293 Runner.Exports < tests/exports.json
