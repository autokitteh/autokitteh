package main

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/runtimes/nodejsrt"
)

func main() {
	var (
		tarFile   = flag.String("tar", "", "Path to the tar file containing the code")
		outputDir = flag.String("output", "", "Output directory for the prepared environment")
	)
	flag.Parse()

	if *tarFile == "" || *outputDir == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -tar <tar-file> -output <output-dir>\n", os.Args[0])
		os.Exit(1)
	}

	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Read tar file
	tarData, err := os.ReadFile(*tarFile)
	if err != nil {
		logger.Error("Failed to read tar file", zap.Error(err))
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		logger.Error("Failed to create output directory", zap.Error(err))
		os.Exit(1)
	}

	// Create a LocalNodeJS instance for preparation
	r := nodejsrt.NewLocalNodeJS(logger)

	// Prepare the project in the specified directory
	if err := r.PrepareProject(tarData, *outputDir); err != nil {
		logger.Error("Failed to prepare project", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Successfully prepared environment",
		zap.String("output_dir", *outputDir))
}
