.PHONY: all
all: protoc

.PHONY: protoc
protoc:
	docker build -t "autokitteh/protoc" --build-arg ARCH="$(shell uname -m)" .
