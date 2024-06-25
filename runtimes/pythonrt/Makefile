ifeq ($(shell uname),Linux)
	open=xdg-open
else
	open=open
endif

export PYTHONPATH=$(PWD)/py-sdk

test: test-py test-go

ci: deps test-py

test-py:
	ruff check .
	ruff format --check
	python -m pytest

# You can set the TESTOPTS to pass options to `go test`
test-go:
	golangci-lint run ./...
	go test $(TESTOPTS) -v

deps:
	python -m pip install -r dev-requirements.txt

# The below targets are for local testing.

deploy-workflow:
	ak deploy -m ./testdata/simple/autokitteh.yaml -d ./testdata/simple/

run-workflow:
	curl -i -X POST -d '{"user": "joe", "event": "login"}' http://localhost:9980/http/py_simple/trigger_url_path_py

deploy-issues:
	ak deploy -m ./testdata/issues/autokitteh.yaml -d ./testdata/issues/

temporal-ui:
	$(open) http://localhost:8233

ngrok:
	ngrok http --domain autokitteh-miki.ngrok.dev 9980

ngrok-prod:
	ngrok http --domain autokitteh.ngrok.dev 9980


cfg_dir = $(shell ak config where  | grep 'Data home' | awk '{print $$NF}')

clean-dbs:
	rm -f $(cfg_dir)/*sql*

test-workflow: deploy-workflow run
