package builds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var project string

var uploadCmd = common.StandardCommand(&cobra.Command{
	Use:     "upload <build file path> <--project=...>",
	Short:   "Upload local build data to server",
	Aliases: []string{"up", "u"},
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		r := resolver.Resolver{Client: common.Client()}
		p, _, err := r.ProjectNameOrID(project)
		if err != nil {
			return err
		}
		if p == nil {
			err = fmt.Errorf("project %q not found", project)
			return common.NewExitCodeError(common.NotFoundExitCode, err)
		}

		b, err := sdktypes.BuildFromProto(&sdktypes.BuildPB{ProjectId: p.ToProto().ProjectId})
		if err != nil {
			return fmt.Errorf("invalid build: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		id, err := builds().Save(ctx, b, data)
		if err != nil {
			return fmt.Errorf("save build: %w", err)
		}

		common.RenderKV("build_id", id)
		return nil
	},
})

func init() {
	// Command-specific flags.
	uploadCmd.Flags().StringVarP(&project, "project", "p", "", "project name or ID")
	kittehs.Must0(uploadCmd.MarkFlagRequired("project"))
}
