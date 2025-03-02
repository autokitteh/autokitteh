ifeq ($(shell uname),Linux)
	open=xdg-open
else
	open=open
endif

export PYTHONPATH=$(PWD)/py-sdk

test: test-py test-go

ci: deps test-py check-sdk-docs-requirements

test-py:
	cd runner && $(MAKE) test

# You can set the TESTOPTS to pass options to `go test`
test-go:
	../../bin/tools/golangci-lint run ./...
	go test -v $(TESTOPTS)

deps:
	cd runner && $(MAKE) install-deps

# The below targets are for local testing.

deploy-simple:
	ak deploy -m ./testdata/simple/autokitteh.yaml -d ./testdata/simple/

run-simple:
	go run ./testdata/trigger_webhook.go

test-simple: deploy-simple run-simple
	ak session log --prints-only --no-timestamps
	sleep 3
	ak session list -j | jq .state

deploy-issues:
	ak deploy -m ./testdata/issues/autokitteh.yaml -d ./testdata/issues/

temporal-ui:
	$(open) http://localhost:8233

ngrok:
	ngrok http --domain autokitteh-${USER}.ngrok.dev 9980

ngrok-prod:
	ngrok http --domain autokitteh.ngrok.dev 9980


cfg_dir = $(shell ak config where  | grep 'Data home' | awk '{print $$NF}')

clean-dbs:
	rm -f $(cfg_dir)/*sql*

test-workflow: deploy-workflow run

check-sdk-docs-requirements:
	cd py-sdk/docs && ./gen-reqs.py
	git add -N .
	git diff --name-only --exit-code
