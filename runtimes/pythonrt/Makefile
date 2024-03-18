ifeq ($(shell uname),Linux)
	open=xdg-open
else
	open=open
endif

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

deploy-workflow:
	go run testdata/deploy.go -- ./testdata/simple/

run-workflow:
	curl -i -X POST http://localhost:9980/http/trigger_url_path_py

temporal-ui:
	$(open) http://localhost:8233

