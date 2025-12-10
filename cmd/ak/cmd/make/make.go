package make

import (
	"maps"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var (
	lint      = []string{"ruff", "check"}
	format    = []string{"ruff", "format", "--check", "."}
	typecheck = []string{"mypy", "--follow-untyped-imports", "."}

	cmds = map[string][][]string{
		"lint":      {lint},
		"format":    {format},
		"typecheck": {typecheck},
		"all":       {lint, format, typecheck},
	}
)

var makeCmd = common.StandardCommand(&cobra.Command{
	Use:   "make",
	Short: "Make: " + strings.Join(slices.Collect(maps.Keys(cmds)), ", "),
	Args:  cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(makeCmd)
}

func init() {
	// Subcommands.
	for name, scripts := range cmds {
		makeCmd.AddCommand(newUVXCmd(name, scripts))
	}
}

func newUVXCmd(name string, script [][]string) *cobra.Command {
	return common.StandardCommand(&cobra.Command{
		Use: name,
		Short: strings.Join(kittehs.Transform(script, func(tool []string) string {
			return strings.Join(tool, " ")
		}), " && "),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			for _, tool := range script {
				args := []string{
					"--with", "autokitteh",
				}

				if f, err := os.Open("requirements.txt"); err == nil {
					f.Close()
					args = append(args, "--with-requirements", "requirements.txt")
				}

				args = append(args, tool...)

				uvx := exec.Command("uvx", args...)

				cmd.Println(uvx.String())

				uvx.Stdout = cmd.OutOrStdout()
				uvx.Stderr = cmd.ErrOrStderr()

				if err := uvx.Run(); err != nil {
					return err
				}
			}

			return nil
		},
	})
}
