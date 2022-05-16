.PHONY: all
all: test

.PHONY: test
test: test-local test-grpc

.PHONY: test-local
test-local:
	set -e; for f in *.star; do ./aksh.sh $${f}; done

.PHONY: test-grpc
test-grpc:
	DSN="grpc:127.0.0.1:50001" make test-local
