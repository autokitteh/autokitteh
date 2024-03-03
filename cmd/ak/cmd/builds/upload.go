package builds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var uploadCmd = common.StandardCommand(&cobra.Command{
	Use:     "upload <build file path>",
	Short:   "Upload local build data to server",
	Aliases: []string{"up", "u"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		id, err := builds().Save(ctx, sdktypes.InvalidBuild, data)
		if err != nil {
			return fmt.Errorf("save build: %w", err)
		}

		common.RenderKV("build_id", id)
		return nil
	},
})
