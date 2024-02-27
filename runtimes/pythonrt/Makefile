all:
	$(error please pick a target)

test-py:
	ruff ak_runner.py
	python -m pytest \
		--doctest-modules \
		--ignore testdata \
		-v .

dev-deps:
	python -m pip install -r dev-requirements.txt
