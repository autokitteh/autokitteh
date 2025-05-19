default:
	$(error please pick a target)

test:
	uv run ruff check .
	uv run ruff format --check
	uv run python -m pytest

install-deps:
	uv venv
	uv pip install .[all]
	rm -rf build *.egg-info

ci: install-deps test

call-start:
	grpcurl --plaintext -d @ localhost:9293 Runner.Start < tests/start.json

call-exports:
	grpcurl --plaintext -d @ localhost:9293 Runner.Exports < tests/exports.json
