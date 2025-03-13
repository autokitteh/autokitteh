package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"go.autokitteh.dev/autokitteh/runtimes/nodejsrt"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

func main() {
	runtimeDir := flag.String("runtime", "", "Path to the runtime directory")
	entryPoint := flag.String("entry-point", "index.js:main", "Entry point in format file:function")
	flag.Parse()

	if *runtimeDir == "" {
		fmt.Println("Please provide runtime directory with -runtime flag")
		os.Exit(1)
	}

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	// Read the tar file
	buildTar := filepath.Join(*runtimeDir, "..", "build", "project_build.tar")
	tarData, err := os.ReadFile(buildTar)
	if err != nil {
		fmt.Printf("Failed to read build tar: %v\n", err)
		os.Exit(1)
	}

	// Create NodeJS runtime with local runner
	cfg := &nodejsrt.Config{
		RunnerType:    "local",
		LogRunnerCode: true,
	}

	// Create runtime
	rt, err := nodejsrt.New(cfg, logger, func() string { return "localhost:0" })
	if err != nil {
		fmt.Printf("Failed to create runtime: %v\n", err)
		os.Exit(1)
	}

	// Create service
	svc, err := rt.New()
	if err != nil {
		fmt.Printf("Failed to create service: %v\n", err)
		os.Exit(1)
	}

	// Create run ID and session ID
	runID := sdktypes.NewRunID()
	sessionID := sdktypes.NewSessionID()

	// Setup callbacks
	cbs := &sdkservices.RunCallbacks{
		Call: func(ctx context.Context, rid sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
			return sdktypes.Nothing, nil
		},
		Load: func(ctx context.Context, rid sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
			return map[string]sdktypes.Value{}, nil
		},
		Print: func(ctx context.Context, rid sdktypes.RunID, text string) error {
			fmt.Println(text)
			return nil
		},
	}

	// Run the code
	ctx := context.Background()
	compiledData := map[string][]byte{
		"code.tar": tarData,
	}

	run, err := svc.Run(ctx, runID, sessionID, *entryPoint, compiledData, nil, cbs)
	if err != nil {
		fmt.Printf("Failed to run code: %v\n", err)
		os.Exit(1)
	}

	// Wait for completion
	if err := run.Stop(ctx); err != nil {
		fmt.Printf("Error during execution: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Execution completed successfully\n")
}
