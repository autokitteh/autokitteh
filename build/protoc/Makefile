ARCH=$(shell uname -m)

.PHONY: all
all: protoc

.PHONY: protoc
protoc:
	@if [[ "$(shell docker version -f json | jq .Server.Experimental)" != "true" ]]; then \
		echo "error: docker experimental mode must be enabled" && exit 1; \
	fi
	docker build --squash -t "protoc-${ARCH}" --build-arg ARCH="${ARCH}" .
