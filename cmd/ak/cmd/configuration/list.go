package configuration

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var envVars bool

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--env_vars]",
	Short:   "List all configurations",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		// It's ok to use dev mode here - the config list is the same
		// for every mode. We use dev so it wont break on missing
		// vars.

		// Initialize the service so that configs would be populated.
		// Configset modes don't matter since all modes share the same keys.
		if _, err := common.NewDevSvc(true); err != nil {
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
	listCmd.Flags().BoolVarP(&envVars, "env-vars", "e", false, "print key names as environment variables (default = false)")
}

// TODO: Fix this to return real values ASAP, see PR #323
func formatConfig(s string) string {
	if envVars {
		s = common.EnvVarPrefix + strings.ToUpper(strings.ReplaceAll(s, ".", "__"))
	}
	return s
}
