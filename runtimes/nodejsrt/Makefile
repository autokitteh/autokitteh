ifeq ($(shell uname),Linux)
	open=xdg-open
else
	open=open
endif

# Set default directories
EXAMPLES_DIR=examples
TESTDATA_DIR=nodejs/testdata
DEFAULT_EXAMPLE=simple-test

# Detect OS for correct open command
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	OPEN_CMD = xdg-open
else
	OPEN_CMD = open
endif

.PHONY: test-nodejs build run execute direct-run clean-nodejs deps-nodejs lint-nodejs

# Default target
all: build run execute

# Run tests for Node.js runtime
test-nodejs:
	cd nodejs && npm test

# Install dependencies for Node.js runtime
deps-nodejs:
	cd nodejs && npm install

# Lint Node.js code
lint-nodejs:
	cd nodejs && npm run lint

# Build target - compile the code and create a tar file
build:
	go run cmd/buildjs/main.go -input=$(EXAMPLES_DIR)/$(DEFAULT_EXAMPLE) -output=$(TESTDATA_DIR)/$(DEFAULT_EXAMPLE)/code.tar -extract=true

# Alternative build with custom input
#build-example:
#	@read -p "Enter example directory name (default: $(DEFAULT_EXAMPLE)): " example; \
#	example=$${example:-$(DEFAULT_EXAMPLE)}; \
#	go run cmd/buildjs/main.go -input=$(EXAMPLES_DIR)/$$example -output=$(TESTDATA_DIR)/$$example/code.tar -extract=true

# Run target - run the code from tar file
deploy:
	go run cmd/deployjs/main.go -tar=$(TESTDATA_DIR)/$(DEFAULT_EXAMPLE)/code.tar -output=$(TESTDATA_DIR)/$(DEFAULT_EXAMPLE)/runtime -setup-deps=true

# Execute target - directly execute the runtime
execute:
	cd nodejs && npx ts-node dev/direct-execution/main.ts --input-dir=../$(TESTDATA_DIR)/$(DEFAULT_EXAMPLE)/runtime

# Directly run with debug output
debug-execute:
	cd nodejs && DEBUG=1 npx ts-node dev/direct-execution/runner_direct.ts $(TESTDATA_DIR)/$(DEFAULT_EXAMPLE)/runtime

# Clean target
clean-nodejs:
	rm -rf $(TESTDATA_DIR)/$(DEFAULT_EXAMPLE)

# Local development workflow - build, run, and execute in sequence
dev: build run execute

# Deploy a test example
deploy-simple:
	ak deploy -m ./examples/$(DEFAULT_EXAMPLE)/autokitteh.yaml -d ./examples/$(DEFAULT_EXAMPLE)/

# Run a webhook trigger
run-webhook:
	go run ./testdata/trigger_webhook.go

# Run the Node.js runner directly (useful for debugging)
direct-run:
	cd nodejs && npx ts-node dev/direct-execution/direct_runner.test.ts

# Show help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the example code and create a tar file"
	@echo "  make deploy            - Run the code from the tar file"
	@echo "  make execute        - Execute the runtime directly"
	@echo "  make dev            - Run build, run, and execute in sequence"
	@echo "  make test-nodejs    - Run Node.js tests"
	@echo "  make deps-nodejs    - Install Node.js dependencies"
	@echo "  make lint-nodejs    - Lint Node.js code"
	@echo "  make clean-nodejs   - Clean output files"
	@echo "  make direct-run     - Run the Node.js runner directly (for debugging)"
	@echo "  make debug-execute  - Run execute with debug output"
	@echo ""
	@echo "Example:"
	@echo "  make dev            - Build, run, and execute default example"
