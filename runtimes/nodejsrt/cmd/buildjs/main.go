package main

import (
	"archive/tar"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"go.autokitteh.dev/autokitteh/runtimes/nodejsrt"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

// BuildConfig holds the configuration for the build process
type BuildConfig struct {
	InputDir    string // Directory containing source code
	OutputDir   string // Directory for build output (optional)
	Environment string // dev/prod environment
	Extract     bool   // Whether to extract the tar file after building
}

// extractTar extracts a tar file to a directory
func extractTar(tarData []byte, destDir string) error {
	tr := tar.NewReader(bytes.NewReader(tarData))

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar: %v", err)
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %v", target, err)
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %v", dir, err)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("error creating file %s: %v", target, err)
			}
			defer f.Close()

			if _, err := io.Copy(f, tr); err != nil {
				return fmt.Errorf("error writing to file %s: %v", target, err)
			}
		}
	}
	return nil
}

func main() {
	// Parse command line flags
	cfg := BuildConfig{}
	flag.StringVar(&cfg.InputDir, "input", "", "Input directory containing source code")
	flag.StringVar(&cfg.OutputDir, "output", "", "Output path for build artifact (optional)")
	flag.StringVar(&cfg.Environment, "env", "dev", "Build environment (dev/prod)")
	flag.BoolVar(&cfg.Extract, "extract", false, "Extract the tar file after building")
	flag.Parse()

	// Validate input
	if cfg.InputDir == "" {
		fmt.Fprintln(os.Stderr, "Error: Input directory is required. Use -input flag to specify the directory")
		os.Exit(1)
	}

	// Setup logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Get absolute path for input directory
	absPath, err := filepath.Abs(cfg.InputDir)
	if err != nil {
		logger.Fatal("Failed to get absolute path",
			zap.String("input", cfg.InputDir),
			zap.Error(err))
	}

	// Create runtime config
	rtConfig := &nodejsrt.Config{
		RunnerType:   "local",
		LogBuildCode: true,
	}

	// Create NodeJS runtime
	logger.Info("Creating NodeJS runtime",
		zap.String("runnerType", rtConfig.RunnerType),
		zap.String("inputDir", absPath))

	rt, err := nodejsrt.New(rtConfig, logger, func() string { return "localhost:0" })
	if err != nil {
		logger.Fatal("Failed to create runtime", zap.Error(err))
	}

	// Create runtime service
	svc, err := rt.New()
	if err != nil {
		logger.Fatal("Failed to create service", zap.Error(err))
	}

	// Run build
	logger.Info("Starting build process")
	artifact, err := svc.(interface {
		Build(context.Context, fs.FS, string, []sdktypes.Symbol) (sdktypes.BuildArtifact, error)
	}).Build(context.Background(), os.DirFS(absPath), "", nil)

	if err != nil {
		logger.Fatal("Build failed", zap.Error(err))
	}

	// Get the tar data
	tarData, ok := artifact.CompiledData()["code.tar"]
	if !ok {
		logger.Fatal("Build artifact does not contain code.tar")
	}

	// Write tar file if output path is specified
	if cfg.OutputDir != "" {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(cfg.OutputDir), 0755); err != nil {
			logger.Fatal("Failed to create output directory",
				zap.String("outputDir", filepath.Dir(cfg.OutputDir)),
				zap.Error(err))
		}

		// Write the tar file
		if err := os.WriteFile(cfg.OutputDir, tarData, 0644); err != nil {
			logger.Fatal("Failed to write tar file",
				zap.String("output", cfg.OutputDir),
				zap.Error(err))
		}
		logger.Info("Wrote build artifact", zap.String("path", cfg.OutputDir))

		// Extract the tar file if requested
		if cfg.Extract {
			extractDir := filepath.Join(filepath.Dir(cfg.OutputDir), "extracted")
			if err := extractTar(tarData, extractDir); err != nil {
				logger.Fatal("Failed to extract tar file",
					zap.String("output", extractDir),
					zap.Error(err))
			}
			logger.Info("Extracted build artifact", zap.String("path", extractDir))
		}
	}

	// Print build summary
	logger.Info("Build completed",
		zap.Bool("isValid", artifact.IsValid()),
		zap.Int("exportCount", len(artifact.Exports())))

	// Print exports
	for _, exp := range artifact.Exports() {
		pb := exp.ToProto()
		logger.Info("Export",
			zap.String("symbol", pb.Symbol),
			zap.String("location", pb.Location.GetPath()),
			zap.Uint32("line", pb.Location.GetRow()))
	}
}
