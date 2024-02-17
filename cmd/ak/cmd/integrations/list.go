package integrations

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	name     string
	withDesc bool
	withRefs bool
	withMod  bool
)

var listCmd = common.StandardCommand(&cobra.Command{
	Use:     "list [--name=...] [--fail] [--with-desc] [--with-refs] [--with-modules]",
	Short:   "List all registered integrations",
	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := common.LimitedContext()
		defer cancel()

		is, err := integrations().List(ctx, name)
		if err != nil {
			return fmt.Errorf("list integrations: %w", err)
		}

		if len(is) == 0 {
			var dummy *sdktypes.Integration
			return common.FailIfNotFound(cmd, "integrations", dummy)
		}

		if !(withDesc && withMod && withRefs) {
			for idx, integ := range is {
				is[idx], err = integ.Update(func(pb *sdktypes.IntegrationPB) {
					if !withDesc {
						pb.Description = ""
					}
					if !withRefs {
						pb.UserLinks = nil
					}
					if !withMod {
						pb.Module = nil
					}
				})
				if err != nil {
					return fmt.Errorf("omit extra details: %w", err)
				}
			}
		}

		common.RenderList(is)
		return nil
	},
})

func init() {
	// Command-specific flags.
	listCmd.Flags().StringVarP(&name, "name", "n", "", "substring in the (ID or display) name")
	listCmd.Flags().BoolVarP(&withDesc, "with-desc", "d", false, "include description")
	listCmd.Flags().BoolVarP(&withRefs, "with-refs", "r", false, "include reference links")
	listCmd.Flags().BoolVarP(&withMod, "with-module", "m", false, "include module details")

	common.AddFailIfNotFoundFlag(listCmd)
}
