package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	cmd = exec.Command("ak", "project", "build", mfst.Project.Name, "--from", dir)
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: can't build - %s\n", err)
		os.Exit(1)
	}

	// build_id: bld_60mfa7qs3927q8j7kc4g4bancx
	out := buf.String()
	i := strings.Index(out, " ")
	if i == -1 {
		fmt.Fprintf(os.Stderr, "error: bad build output - %q\n", out)
		os.Exit(1)
	}
	buildID := buf.String()[i+1 : len(out)-1] // ignore \n

	buf.Reset()

	cmd = exec.Command("ak", "env", "list")
	cmd.Stdout = &buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: can't get env list - %s\n", err)
		os.Exit(1)
	}

	// env_id:"env_60j2xb5s3927q8j7kc4g4bancx"  project_id:"prj_60h3snbs3927q8j7kc4g4bancx"  name:"default"
	envID := ""
	s := bufio.NewScanner(&buf)
	for s.Scan() {
		if strings.Contains(s.Text(), `name:"default"`) {
			i := strings.Index(s.Text(), `"`)
			if i == -1 {
				fmt.Fprintf(os.Stderr, "error: bad env line - %q\n", s.Text())
				os.Exit(1)
			}

			j := strings.Index(s.Text()[i+1:], `"`)
			if j == -1 {
				fmt.Fprintf(os.Stderr, "error: bad env line - %q\n", s.Text())
				os.Exit(1)
			}
			envID = s.Text()[i+1 : i+j+1]
		}
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: can't scan env list - %s\n", err)
		os.Exit(1)
	}

	if envID == "" {
		fmt.Fprintln(os.Stderr, "error: can't find default env ID")
		os.Exit(1)
	}

	cmd = exec.Command("ak", "deployment", "create", "--build-id", buildID, "--env", envID, "--activate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: can't create deployment - %s\n", err)
		os.Exit(1)
	}
}
