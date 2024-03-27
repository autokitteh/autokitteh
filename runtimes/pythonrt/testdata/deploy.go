package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"gopkg.in/yaml.v3"
)

func defaultEnvID() (string, error) {
	var buf bytes.Buffer

	cmd := exec.Command("ak", "-j", "env", "list")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("env list - %w", err)
	}

	dec := json.NewDecoder(&buf)
	var env struct {
		ID   string `json:"env_id"`
		Name string
	}

	for {
		err := dec.Decode(&env)
		if errors.Is(err, io.EOF) {
			return "", fmt.Errorf("can't find default env in env list")
		}

		if err != nil {
			return "", fmt.Errorf("env list - %w", err)
		}

		if env.Name == "default" {
			return env.ID, nil
		}
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s DIR\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "error: wrong number of arguments")
		os.Exit(1)
	}

	dir := flag.Arg(0)
	mfstFile := path.Join(dir, "autokitteh.yaml")
	file, err := os.Open(mfstFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var mfst struct {
		Project struct {
			Name string
		}
	}
	if err := yaml.NewDecoder(file).Decode(&mfst); err != nil {
		fmt.Fprintf(os.Stderr, "error: %q - %s\n", mfstFile, err)
		os.Exit(1)
	}

	cmd := exec.Command("ak", "manifest", "apply", mfstFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: can't apply manifest - %s\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	cmd = exec.Command("ak", "-j", "project", "build", mfst.Project.Name, "--from", dir)
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: can't build - %s\n", err)
		os.Exit(1)
	}

	var build struct {
		ID string `json:"build_id"`
	}

	if err := json.NewDecoder(&buf).Decode(&build); err != nil {
		fmt.Fprintf(os.Stderr, "error: bad build output - %q\n", buf.String())
		os.Exit(1)
	}
	fmt.Println("build ID:", build.ID)

	envID, err := defaultEnvID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: can't find default env - %s\n", err)
		os.Exit(1)
	}

	cmd = exec.Command("ak", "deployment", "create", "--build-id", build.ID, "--env", envID, "--activate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: can't create deployment - %s\n", err)
		os.Exit(1)
	}
}
