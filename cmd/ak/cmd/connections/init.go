package connections

import (
	"errors"
	"net/url"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var initCmd = common.StandardCommand(&cobra.Command{
	Use:     "init <connection name or ID>",
	Short:   "Initialize connection",
	Aliases: []string{"i"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		r := resolver.Resolver{Client: common.Client()}
		ctx, cancel := common.LimitedContext()
		defer cancel()

		c, _, err := r.ConnectionNameOrID(ctx, args[0], "")
		if err = common.AddNotFoundErrIfCond(err, c.IsValid()); err != nil {
			return common.ToExitCodeWithSkipNotFoundFlag(cmd, err, "connection")
		}

		link := kittehs.Must1(url.JoinPath(c.Links().InitURL(), "cli"))
		if link == "" {
			return errors.New("connection doesn't have an init link")
		}

		return common.OpenURL(cmd, link)
	},
})

func init() {
	// Command-specific flags.
	common.AddFailIfNotFoundFlag(initCmd)
}
