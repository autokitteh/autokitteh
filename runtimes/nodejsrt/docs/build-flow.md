# NodeJS Runtime Build Flow

This diagram shows the build process flow for the NodeJS runtime, specifically the interaction between `build.go` and `build-patch.ts`:

```mermaid
sequenceDiagram
    participant Client as Client (ak deploy)
    participant Deploy as Server (Deploy Service)
    participant Build as Server (Build Service)
    participant NodeBuild as NodeJS RT (build.go)
    participant BuildPatch as NodeJS RT (build-patch.ts)

    Client->>Deploy: ak deploy (project files)
    Deploy->>Build: BuildProject()
    Build->>NodeBuild: Build(fsys, path)
    
    %% build.go process
    NodeBuild->>NodeBuild: Filter files (skip node_modules)
    
    %% Interaction with build-patch.ts
    NodeBuild->>BuildPatch: Process source files
    BuildPatch->>BuildPatch: Parse JS/TS
    BuildPatch->>BuildPatch: Inject ak_call wrappers
    BuildPatch-->>NodeBuild: Return patched code
    
    %% Back to build.go
    NodeBuild->>NodeBuild: Find exports
    NodeBuild->>NodeBuild: Create tar with patched code
    
    NodeBuild-->>Build: Return BuildArtifact
    Build-->>Deploy: Return build ID
    Deploy->>Deploy: Create deployment record
    Deploy-->>Client: Deployment complete
```

## Components

- `build.go`: Handles the overall build process including file filtering, export discovery, and artifact creation
- `build-patch.ts`: Responsible for code transformation, specifically injecting `ak_call` wrappers around function calls