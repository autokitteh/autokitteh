package export

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Flags shared by the "create" and "list" subcommands.
var env, connection string

var Cmd = common.StandardCommand(&cobra.Command{
	Use:     "export <project name or ID>",
	Short:   "Export project",
	Aliases: []string{"ex"},
	Args:    cobra.ExactArgs(1),

	RunE: export,
})

func export(cmd *cobra.Command, args []string) error {
	r := resolver.Resolver{Client: common.Client()}
	ctx, cancel := common.LimitedContext()
	defer cancel()

	name := args[0]
	prj, _, err := r.ProjectNameOrID(ctx, name)
	if err != nil {
		return err
	}

	if !prj.IsValid() {
		return fmt.Errorf("project %q not found", name)
	}

	tf := sdkservices.ListTriggersFilter{
		ProjectID: prj.ID(),
	}
	triggers, err := r.Client.Triggers().List(ctx, tf)
	if err != nil {
		return err
	}

	fmt.Println("version: v1")
	fmt.Println()
	fmt.Println("project:")
	fmt.Printf("  name: %s\n", prj.Name().String())
	fmt.Println("  triggers:")

	for _, t := range triggers {
		fmt.Printf("    - name: %s\n", t.Name().String())
		fmt.Printf("      event_type: %s\n", t.EventType())
		fmt.Printf("      call: %s\n", t.CodeLocation().CanonicalString())
		switch t.SourceType() {
		case sdktypes.TriggerSourceTypeWebhook:
			fmt.Println("      webhook: {}")
			// TODO: More types
		}
	}

	envs, err := r.Client.Envs().List(ctx, prj.ID())
	if err != nil {
		return err
	}

	fmt.Println("  vars:")
	for _, env := range envs {
		sid, err := sdktypes.Strict(sdktypes.ParseVarScopeID(env.ID().String()))
		if err != nil {
			return err
		}
		vars, err := r.Client.Vars().Get(ctx, sid)
		for _, v := range vars {
			if v.IsSecret() {
				continue
			}
			fmt.Printf("    - name: %s\n", v.Name().String())
			fmt.Printf("      value: %s\n", v.Value())
		}
	}

	return nil
}
