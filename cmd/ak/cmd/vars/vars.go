package vars

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Flags shared by all the subcommands.
var project, conn string

var varsCmd = common.StandardCommand(&cobra.Command{
	Use:   "var",
	Short: "Connection variable subcommands: set, get, delete",
	Args:  cobra.NoArgs,
})

// AddSubcommands adds this command, and its own subcommands, to the calling parent.
func AddSubcommands(parentCmd *cobra.Command) {
	parentCmd.AddCommand(varsCmd)
}

func init() {
	// Flag shared by all subcommands.
	// We don't define it as a single persistent flag here
	// because then we wouldn't be able to mark it as required.
	getCmd.Flags().StringVarP(&conn, "connection", "c", "", "connection name or ID")
	setCmd.Flags().StringVarP(&conn, "connection", "c", "", "connection name or ID")
	deleteCmd.Flags().StringVarP(&conn, "connection", "c", "", "connection name or ID")

	// Flag shared by all subcommands.
	// We don't define it as a single persistent flag here for aesthetic
	// conformance with the "env" flag, and "project" in other "envs" sibling commands.
	getCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	setCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	deleteCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")

	// Subcommands.
	varsCmd.AddCommand(setCmd)
	varsCmd.AddCommand(getCmd)
	varsCmd.AddCommand(deleteCmd)
}

func vars() sdkservices.Vars {
	return common.Client().Vars()
}

func resolveScopeID() (sdktypes.VarScopeID, error) {
	r := resolver.Resolver{Client: common.Client()}

	ctx, cancel := common.LimitedContext()
	defer cancel()

	if conn != "" {
		c, id, err := r.ConnectionNameOrID(ctx, conn, project)
		if err != nil {
			return sdktypes.InvalidVarScopeID, err
		}
		if !c.IsValid() {
			err = fmt.Errorf("connection %q not found", conn)
			return sdktypes.InvalidVarScopeID, common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		return sdktypes.NewVarScopeID(id), nil
	}

	p, id, err := r.ProjectNameOrID(ctx, project)
	if err != nil {
		return sdktypes.InvalidVarScopeID, err
	}
	if !p.IsValid() {
		err = fmt.Errorf("project %q not found", project)
		return sdktypes.InvalidVarScopeID, common.NewExitCodeError(common.NotFoundExitCode, err)
	}

	return sdktypes.NewVarScopeID(id), nil
}
