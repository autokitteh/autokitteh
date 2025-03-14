#!/bin/bash

# Exit on any error
set -e

# Get the absolute path to the tools directory
TOOLS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Build all tools
echo "Building tools..."
cd "$TOOLS_DIR/build" && go build -o build_direct build_direct.go
cd "$TOOLS_DIR/runner" && go build -o runner_direct runner_direct.go
cd "$TOOLS_DIR/execute" && go build -o execute_direct execute_direct.go
cd "$TOOLS_DIR/prepare" && go build -o prepare_test_env prepare_test_env.go

# If a project directory was provided, run prepare_test_env
if [ $# -ge 1 ]; then
    echo "Preparing test environment for: $1"
    ENTRY_POINT=${2:-"index.js:main"}  # Default to index.js:main if not provided
    
    # Prepare the environment
    cd "$TOOLS_DIR/prepare" && ./prepare_test_env -project "$1" -entry-point "$ENTRY_POINT"
    
    # Get the runtime directory
    PROJECT_NAME=$(basename "$1")
    RUNTIME_DIR="$TOOLS_DIR/test_data/prepared/$PROJECT_NAME/runtime"
    
    # Execute the code
    echo "Executing code..."
    cd "$TOOLS_DIR/execute" && ./execute_direct -runtime "$RUNTIME_DIR" -entry-point "$ENTRY_POINT"
fi

echo "Done!" 