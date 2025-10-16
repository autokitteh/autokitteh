# NodeJS Runtime Run Flow

This diagram shows the runtime execution flow for the NodeJS runtime, detailing how `ak run` command executes code and handles activities through the Node.js runner:

```mermaid
sequenceDiagram
    participant Client as Client (ak run)
    participant Runtime as Server (Runtime Service)
    participant Store as Server (Artifact Store)
    participant Manager as NodeJS RT (runner_manager.go)
    participant Runner as NodeJS RT (local_runner.go)
    participant NodeProcess as NodeJS Process (main.ts)
    participant UserCode as User Code
    participant Activities as Server (Activities)

    Client->>Runtime: ak run --project myproj --function handler
    Runtime->>Store: Get deployment artifact
    Store-->>Runtime: Return project tar
    
    %% Runner manager starts the process
    Runtime->>Manager: StartRunner(artifact, env)
    Manager->>Runner: Start()
    Runner->>Runner: Extract code
    Runner->>Runner: Setup environment
    Runner->>NodeProcess: Start Node.js process<br/>(main.ts with args)
    
    %% main.ts initialization
    NodeProcess->>NodeProcess: Start gRPC server
    NodeProcess->>NodeProcess: Load user code
    NodeProcess->>UserCode: Import and execute function
    
    %% ak_call flow
    UserCode->>NodeProcess: External function call
    Note over UserCode,NodeProcess: Wrapped in ak_call by<br/>build-patch.ts during build
    NodeProcess->>Activities: Activity request
    Note over NodeProcess,Activities: gRPC: ExecuteActivity
    Activities->>Activities: Execute function
    Activities-->>NodeProcess: Activity result
    NodeProcess-->>UserCode: Return result
    
    %% Completion
    UserCode-->>NodeProcess: Function complete
    NodeProcess-->>Runner: Execution complete
    Runner-->>Manager: Runner complete
    Manager-->>Runtime: Return result
    Runtime-->>Client: Return execution result
```

## Components

- `Client (ak run)`: Command line interface for executing project functions
- `Server (Runtime Service)`: Manages execution sessions and coordinates with runners
- `Server (Artifact Store)`: Stores and provides access to deployment artifacts
- `NodeJS RT (runner_manager.go)`: Manages Node.js runner instances
- `NodeJS RT (local_runner.go)`: Handles local execution of Node.js code
- `NodeJS Process (main.ts)`: The actual Node.js runner process that executes code
- `User Code`: The deployed project code being executed
- `Server (Activities)`: Handles execution of external function calls

## Flow Description

1. **Execution Initiation**
   - User runs `ak run --project myproj --function handler`
   - Runtime service retrieves deployment artifact
   - Artifact is extracted to temporary directory

2. **Runner Setup**
   - Runner manager creates new runner instance
   - Local runner extracts code and sets up environment
   - `main.ts` is started with necessary arguments:
     - `--worker-address`: gRPC server address
     - `--port`: Runner's port
     - `--runner-id`: Unique runner ID
     - `--code-dir`: Path to extracted code

3. **Code Execution**
   - Node.js process loads and executes specified function
   - External function calls are intercepted by `ak_call` wrapper
   - Activities are executed through gRPC communication
   - Results are returned to user code

4. **Completion**
   - Function execution completes
   - Results are returned through the chain
   - Resources are cleaned up 