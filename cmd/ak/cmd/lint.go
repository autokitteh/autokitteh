// This is currently a stand alone command, later it'll integrate in the server
package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
)

var (
	checks   []check
	lintOpts struct {
		manifestPath string
	}
)

func init() {
	// Command-specific flags.
	lintCmd.Flags().StringVarP(&lintOpts.manifestPath, "manifest", "m", "", "YAML manifest file containing project settings")

	checks = []check{
		checkNoTriggers,
		checkEmptyVars,
	}
}

var lintCmd = common.StandardCommand(&cobra.Command{
	Use:   "lint",
	Short: "Lint a project",
	Args:  cobra.NoArgs,
	RunE:  runLint,
})

func runLint(cmd *cobra.Command, args []string) error {
	if lintOpts.manifestPath == "" {
		lintOpts.manifestPath = "autokitteh.yaml"
	}

	file, err := os.Open(lintOpts.manifestPath)
	if err != nil {
		return err
	}
	defer file.Close()

	const maxSize = 1 << 20 // 1MB
	data, err := io.ReadAll(io.LimitReader(file, maxSize))
	if err != nil {
		return err
	}

	m, err := manifest.Read(data, lintOpts.manifestPath)
	if err != nil {
		return err
	}

	hasErrors := false
	for _, c := range checks {
		if err = errors.Join(err, c(m, cmd.ErrOrStderr())); err != nil {
			hasErrors = true
		}
	}

	if hasErrors {
		return ErrLint
	}

	return nil
}

type check func(m *manifest.Manifest, out io.Writer) error

var ErrLint = errors.New("lint error")

func checkNoTriggers(m *manifest.Manifest, out io.Writer) error {
	if len(m.Project.Triggers) == 0 {
		fmt.Fprintf(out, "lint: error: no triggers\n")
		return ErrLint
	}

	return nil
}

func checkEmptyVars(m *manifest.Manifest, out io.Writer) error {
	for _, v := range m.Project.Vars {
		fmt.Printf("%s: %q\n", v.Name, v.Value)
		if v.Value == "" {
			fmt.Fprintf(out, "lint: warning: varible %s is empty\n", v.Name)
		}
	}

	return nil
}
