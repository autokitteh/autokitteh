package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	projectDir := flag.String("project", "", "Project directory to prepare")
	entryPoint := flag.String("entry-point", "index.js:main", "Entry point in format file:function")
	flag.Parse()

	if *projectDir == "" {
		fmt.Println("Please provide project directory with -project flag")
		os.Exit(1)
	}

	// Get absolute paths
	absProjectDir, err := filepath.Abs(*projectDir)
	if err != nil {
		fmt.Printf("Failed to get absolute project path: %v\n", err)
		os.Exit(1)
	}

	// Get absolute path to prepare_test_env executable
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	// Get tools directory (parent of prepare/)
	toolsDir := filepath.Dir(filepath.Dir(execPath))

	// Get absolute paths to other executables
	buildExec := filepath.Join(toolsDir, "build", "build_direct")
	runnerExec := filepath.Join(toolsDir, "runner", "runner_direct")

	// Prepare output directory under test_data/prepared/
	projectName := filepath.Base(absProjectDir)
	preparedDir := filepath.Join(toolsDir, "test_data", "prepared", projectName)

	// Create directories
	buildDir := filepath.Join(preparedDir, "build")
	unpackedDir := filepath.Join(preparedDir, "unpacked")
	runtimeDir := filepath.Join(preparedDir, "runtime")

	for _, dir := range []string{buildDir, unpackedDir, runtimeDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	fmt.Println("Phase 1: Building project...")
	buildOutput := filepath.Join(buildDir, "project_build.tar")
	buildCmd := exec.Command(
		buildExec,
		"-input", absProjectDir,
		"-output", buildOutput,
	)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Build failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Phase 2: Unpacking build artifacts...")
	unpackCmd := exec.Command(
		"tar",
		"-xf", buildOutput,
		"-C", unpackedDir,
	)
	unpackCmd.Stdout = os.Stdout
	unpackCmd.Stderr = os.Stderr
	if err := unpackCmd.Run(); err != nil {
		fmt.Printf("Unpack failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Phase 3: Setting up runtime environment...")
	runnerCmd := exec.Command(
		runnerExec,
		"--project", unpackedDir,
		"--output", runtimeDir,
		"--entry-point", *entryPoint,
	)
	runnerCmd.Stdout = os.Stdout
	runnerCmd.Stderr = os.Stderr
	if err := runnerCmd.Run(); err != nil {
		fmt.Printf("Runner setup failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test environment prepared at: %s\n", preparedDir)
}
