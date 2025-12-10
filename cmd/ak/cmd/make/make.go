package make

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
			// Read Makefile from config dir if exists.
			configMakefile := filepath.Join(xdg.ConfigHomeDir(), "Makefile")
			if _, err := os.Stat(configMakefile); err == nil {
				cfgMakefile, err := os.ReadFile(configMakefile)
				if err != nil {
					return err
				}
				makefileData = fmt.Sprintf("# Taken from %s\n\n%s", configMakefile, string(cfgMakefile))
			} else {
				makefileData = fmt.Sprintf("# Using built-in Makefile; create %s to override.\n\n%s", configMakefile, makefileData)
			}

			if show {
				_, err := cmd.OutOrStdout().Write([]byte(makefileData))
				return err
			}

			makefile, err := os.CreateTemp("", "ak-makefile-*.mk")
			if err != nil {
				return err
			}

			defer os.Remove(makefile.Name())
			defer makefile.Close()

			if _, err := makefile.Write([]byte(makefileData)); err != nil {
				return err
			}

			makeArgs := []string{"-f", makefile.Name()}
			makeArgs = append(makeArgs, args...)

			make := exec.Command("make", makeArgs...)

			make.Stdout = cmd.OutOrStdout()
			make.Stderr = cmd.ErrOrStderr()
			make.Stdin = cmd.InOrStdin()

			return make.Run()
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
