package vars

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Flags shared by all the subcommands.
var env, project, conn string

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

	getCmd.Flags().VarP(common.NewNonEmptyString("", &env), "env", "e", "environment name or ID")
	setCmd.Flags().VarP(common.NewNonEmptyString("", &env), "env", "e", "environment name or ID")
	deleteCmd.Flags().VarP(common.NewNonEmptyString("", &env), "env", "e", "environment name or ID")

	getCmd.Flags().VarP(common.NewNonEmptyString("", &conn), "connection", "c", "connection name or ID")
	setCmd.Flags().VarP(common.NewNonEmptyString("", &conn), "connection", "c", "connection name or ID")
	deleteCmd.Flags().VarP(common.NewNonEmptyString("", &conn), "connection", "c", "connection name or ID")

	getCmd.MarkFlagsMutuallyExclusive("env", "connection")
	setCmd.MarkFlagsMutuallyExclusive("env", "connection")
	deleteCmd.MarkFlagsMutuallyExclusive("env", "connection")

	// although either "env" or "connection" is required, if nothing is provided, then
	// var will be resolved from the default env (e.g. env="").
	// So `env' and `connection` are not marked with OneRequiredFlag here.

	// Flag shared by all subcommands.
	// We don't define it as a single persistent flag here for aesthetic
	// conformance with the "env" flag, and "project" in other "envs" sibling commands.
	getCmd.Flags().VarP(common.NewNonEmptyString("", &project), "project", "p", "project name or ID")
	setCmd.Flags().VarP(common.NewNonEmptyString("", &project), "project", "p", "project name or ID")
	deleteCmd.Flags().VarP(common.NewNonEmptyString("", &project), "project", "p", "project name or ID")

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

	if project != "" {
		p, _, err := r.ProjectNameOrID(ctx, project)
		if err = common.AddNotFoundErrIfCond(err, p.IsValid()); err != nil {
			return sdktypes.InvalidVarScopeID, common.ToExitCodeErrorNotNilErr(err, "project")
		}
	}

	if conn != "" {
		c, id, err := r.ConnectionNameOrID(ctx, conn, project)
		if err = common.AddNotFoundErrIfCond(err, c.IsValid()); err != nil {
			return sdktypes.InvalidVarScopeID, common.ToExitCodeErrorNotNilErr(err, "connection")
		}
		return sdktypes.NewVarScopeID(id), nil
	}

	e, id, err := r.EnvNameOrID(ctx, env, project)
	if err = common.AddNotFoundErrIfCond(err, e.IsValid()); err != nil {
		return sdktypes.InvalidVarScopeID, common.ToExitCodeErrorNotNilErr(err, "environment")
	}
	return sdktypes.NewVarScopeID(id), nil
}
