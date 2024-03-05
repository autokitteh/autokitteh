all:
	$(error please pick a target)


ci: deps test

test:
	ruff ak_runner.py
	python -m pytest \
		--doctest-modules \
		--ignore testdata \
		-v .

deps:
	python -m pip install -r dev-requirements.txt
