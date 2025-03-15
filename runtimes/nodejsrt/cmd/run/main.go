package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"go.autokitteh.dev/autokitteh/runtimes/nodejsrt"
	"go.uber.org/zap"
)

type RunConfig struct {
	TarFile   string // Path to the tar file
	OutputDir string // Directory where the environment will be placed
	SetupDeps bool   // Whether to run setupDependencies
	Verbose   bool   // Enable verbose logging
}

func main() {
	// Parse command line flags
	cfg := RunConfig{}
	flag.StringVar(&cfg.TarFile, "tar", "", "Path to the tar file to run")
	flag.StringVar(&cfg.OutputDir, "output", "", "Directory where the environment will be placed")
	flag.BoolVar(&cfg.SetupDeps, "setup-deps", false, "Run setupDependencies to install TypeScript and npm dependencies")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	// Validate input
	if cfg.TarFile == "" {
		fmt.Fprintln(os.Stderr, "Error: Tar file is required. Use -tar flag to specify the file")
		os.Exit(1)
	}
	if cfg.OutputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: Output directory is required. Use -output flag to specify the directory")
		os.Exit(1)
	}

	// Setup logger
	loggerConfig := zap.NewDevelopmentConfig()
	if cfg.Verbose {
		loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger, err := loggerConfig.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Read tar file
	logger.Info("Reading tar file", zap.String("file", cfg.TarFile))
	tarData, err := os.ReadFile(cfg.TarFile)
	if err != nil {
		logger.Fatal("Failed to read tar file",
			zap.String("tar", cfg.TarFile),
			zap.Error(err))
	}

	// Create output directory if it doesn't exist
	absOutputDir, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		logger.Fatal("Failed to get absolute path for output directory",
			zap.String("output", cfg.OutputDir),
			zap.Error(err))
	}
	if err := os.MkdirAll(absOutputDir, 0755); err != nil {
		logger.Fatal("Failed to create output directory",
			zap.String("output", absOutputDir),
			zap.Error(err))
	}

	// Create runner
	logger.Info("Creating NodeJS runner")
	runner := nodejsrt.NewLocalNodeJS(logger)
	defer func() {
		// Don't cleanup since we want to keep the files for inspection
		logger.Info("Runner closed - environment files preserved in output directory",
			zap.String("dir", absOutputDir))
	}()

	// Set project directory and prepare it
	runner.SetProjectDir(absOutputDir)
	logger.Info("Preparing environment", zap.String("dir", absOutputDir))
	if err := runner.PrepareProject(tarData); err != nil {
		logger.Fatal("Failed to prepare environment", zap.Error(err))
	}
	logger.Info("Environment prepared successfully")

	// Setup dependencies if requested
	if cfg.SetupDeps {
		logger.Info("Setting up dependencies")
		if err := runner.SetupDependencies(); err != nil {
			logger.Fatal("Failed to setup dependencies", zap.Error(err))
		}
		logger.Info("Dependencies setup completed")
	}

	// Print instructions for manual inspection
	fmt.Printf("\nEnvironment prepared successfully!\n")
	fmt.Printf("Environment location: %s\n", absOutputDir)
	if cfg.SetupDeps {
		fmt.Printf("Dependencies have been installed.\n")
	}
	fmt.Printf("You can inspect the contents of this directory to verify the setup.\n")
	fmt.Printf("The files will remain in this directory for inspection.\n")
	fmt.Printf("Press Ctrl+C to exit.\n\n")

	// Wait for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}
