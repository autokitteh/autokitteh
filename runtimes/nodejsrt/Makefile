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

clean-temporal:
	killall -9 "temporal-cli-go-sdk-1.31.0"
	rm "/Users/adiludmer/Library/Application Support/autokitteh/temporal_dev.sqlite"
	rm "/Users/adiludmer/Library/Application Support/autokitteh/autokitteh.sqlite"
	rm -rf "/Users/adiludmer/GolandProjects/autokitteh/runtimes/nodejsrt/testdata/simple/dist"

webhook = $(shell ../../bin/ak trigger list | grep -o 'webhook_slug:"[^"]*"' | awk -F'"' '{print $2}' | cut -d '"' -f 2)

start-workflow:
	curl -d'{"user": "joe", "action": "read"}' http://localhost:9980/webhooks/$(webhook)