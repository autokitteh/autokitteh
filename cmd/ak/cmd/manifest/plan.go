package manifest

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	noValidate, fromScratch bool
	org                     string
)

var planCmd = common.StandardCommand(&cobra.Command{
	Use:     "plan [file] [--project-name <name>] [--org org] [--no-validate] [--from-scratch] [--quiet] [--rm-unused-cvars]",
	Short:   "Dry-run for applying a YAML manifest, from a file or stdin",
	Aliases: []string{"p"},
	Args:    cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		r := resolver.Resolver{Client: common.Client()}
		oid, err := r.Org(ctx, org)
		if err != nil {
			return err
		}

		data, path, err := common.Consume(args)
		if err != nil {
			return err
		}

		actions, err := plan(cmd, data, path, projectName, oid)
		if err != nil {
			return err
		}

		common.Render(actions)

		return nil
	},
})

func init() {
	planCmd.Flags().BoolVar(&noValidate, "no-validate", false, "do not validate")
	planCmd.Flags().BoolVarP(&fromScratch, "from-scratch", "s", false, "assume no existing setup")
	planCmd.Flags().BoolVar(&rmUnusedConnVars, "rm-unused-cvars", false, "delete connection variables not used")
	planCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "only show errors, if any")
	planCmd.Flags().StringVarP(&projectName, "project-name", "n", "", "project name")
	planCmd.Flags().StringVarP(&org, "org", "o", "", "org name or id")
}

func plan(cmd *cobra.Command, data []byte, path, projectName string, oid sdktypes.OrgID) (manifest.Actions, error) {
	m, err := manifest.Read(data, path)
	if err != nil {
		return nil, err
	}

	ctx, cancel := common.LimitedContext()
	defer cancel()

	return manifest.Plan(
		ctx, m, common.Client(),
		manifest.WithLogger(logFunc(cmd, "plan")),
		manifest.WithFromScratch(fromScratch),
		manifest.WithProjectName(projectName),
		manifest.WithRemoveUnusedConnFlags(rmUnusedConnVars),
		manifest.WithOrgID(oid),
	)
}
