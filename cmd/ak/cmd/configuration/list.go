package configuration

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/backend/svc"
	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var envVars, showValues bool

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--env_vars] [--values]",
	Short:   "List all configurations",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Fix this to return real values ASAP, see PR #323
		if showValues {
			return fmt.Errorf("--values is not implemented yet")
		}

		// Initialize the service so that configs would be populated.
		// Configset modes don't matter since all modes share the same keys.
		if _, err := svc.New(common.Config(), svc.RunOptions{}); err != nil {
			return err
		}

		// TODO: Fix this to return real values ASAP, see PR #323
		cs := kittehs.Transform(common.Config().ListAll(), formatConfig)
		sort.Strings(cs)
		fmt.Println(strings.Join(cs, "\n"))

		return nil
	},
})

func init() {
	listCmd.Flags().BoolVarP(&envVars, "env_vars", "e", false, "print key names as environment variables (default = false)")
	listCmd.Flags().BoolVarP(&showValues, "values", "v", false, "print default values too - WARNING: may be sensitive! (default = false)")
}

// TODO: Fix this to return real values ASAP, see PR #323
func formatConfig(s string) string {
	if envVars {
		s = common.EnvVarPrefix + strings.ToUpper(strings.ReplaceAll(s, ".", "__"))
	}
	return s
}
