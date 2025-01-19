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

var uploadCmd = common.StandardCommand(&cobra.Command{
	Use:     "upload <build file path> --project <project id>",
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

		r := resolver.Resolver{Client: common.Client()}
		pid, err := r.ProjectNameOrID(ctx, project)
		if err != nil {
			return err
		}

		id, err := builds().Save(ctx, sdktypes.NewBuild().WithProjectID(pid), data)
		if err != nil {
			return fmt.Errorf("save build: %w", err)
		}

		common.RenderKV("build_id", id)
		return nil
	},
})

func init() {
	uploadCmd.Flags().StringVarP(&project, "project", "p", "", "Project ID")
	kittehs.Must0(uploadCmd.MarkFlagRequired("project"))
}
