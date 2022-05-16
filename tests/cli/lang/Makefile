CLITEST=../../../bin/tools/clitest --diff-options '-u -w'

tests=*.clitest

all: test

clean:
	rm -f *.akm

.PHONY: test-grpc
test-grpc: clean
	test_grpc=1 ${CLITEST} $(tests)

.PHONY: test-local
test-local: clean
	${CLITEST} $(tests)

.PHONY: test
test: test-local test-grpc
