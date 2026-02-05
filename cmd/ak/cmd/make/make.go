package makecmd

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/xdg"
)

//go:embed Makefile
var makefileData string

var (
	show bool

	makeCmd = common.StandardCommand(&cobra.Command{
		Use:   "make",
		Short: "Invoke the builtin standard makefile",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			modifiedMakefileData := makefileData

			// Read Makefile from config dir if exists.
			configMakefile := filepath.Join(xdg.ConfigHomeDir(), "Makefile")
			if _, err := os.Stat(configMakefile); err == nil {
				cfgMakefile, err := os.ReadFile(configMakefile)
				if err != nil {
					return err
				}
				modifiedMakefileData = fmt.Sprintf("# Taken from %s\n\n%s", configMakefile, string(cfgMakefile))
			} else {
				modifiedMakefileData = fmt.Sprintf("# Using built-in Makefile; create %s to override.\n\n%s", configMakefile, modifiedMakefileData)
			}

			if show {
				_, err := cmd.OutOrStdout().Write([]byte(modifiedMakefileData))
				return err
			}

			makefile, err := os.CreateTemp("", "ak-makefile-*.mk")
			if err != nil {
				return err
			}

			defer os.Remove(makefile.Name())

			if _, err := makefile.Write([]byte(modifiedMakefileData)); err != nil {
				return err
			}

			if err := makefile.Sync(); err != nil {
				_ = makefile.Close()
				return err
			}

			if err := makefile.Close(); err != nil {
				return err
			}

			makeArgs := []string{"-f", makefile.Name()}
			makeArgs = append(makeArgs, args...)

			make := exec.Command("make", makeArgs...)

			make.Stdout = cmd.OutOrStdout()
			make.Stderr = cmd.ErrOrStderr()
			make.Stdin = cmd.InOrStdin()

			if err := make.Run(); err != nil {
				return fmt.Errorf("failed to execute make command (ensure 'make' is installed and in your PATH): %w", err)
			}
			return nil
		},
	})
)

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(makeCmd)
}

func init() {
	makeCmd.Flags().BoolVarP(&show, "show", "s", false, "Show the makefile instead of executing it")
}
