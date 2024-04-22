package vars

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// Flags shared by all the subcommands.
var env, project string

var varsCmd = common.StandardCommand(&cobra.Command{
	Use:     "vars",
	Short:   "Environment variable subcommands",
	Aliases: []string{"var", "v"},
	Args:    cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(varsCmd)
}

func init() {
	// Flag shared by all subcommands.
	// We don't define it as a single persistent flag here
	// because then we wouldn't be able to mark it as required.
	getCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	kittehs.Must0(getCmd.MarkFlagRequired("env"))

	revealCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	kittehs.Must0(revealCmd.MarkFlagRequired("env"))

	setCmd.Flags().StringVarP(&env, "env", "e", "", "environment name or ID")
	kittehs.Must0(setCmd.MarkFlagRequired("env"))

	// Flag shared by all subcommands.
	// We don't define it as a single persistent flag here for aesthetic
	// conformance with the "env" flag, and "project" in other "envs" sibling commands.
	getCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	revealCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	setCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")

	// Subcommands.
	varsCmd.AddCommand(setCmd)
	varsCmd.AddCommand(getCmd)
	varsCmd.AddCommand(revealCmd)
	varsCmd.AddCommand(removeCmd)
}

func envs() sdkservices.Envs {
	return common.Client().Envs()
}
