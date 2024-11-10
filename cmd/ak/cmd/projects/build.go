package projects

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var buildCmd = common.StandardCommand(&cobra.Command{
	Use:   "build <project name or ID> [--dir <path> [...]] [--file <path> [...]]",
	Short: "Build project",
	Long:  `Build project - see also the "build" parent command`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		id, _, err := common.BuildProject(args[0], dirPaths, filePaths)
		if err == nil {
			common.RenderKVIfV("build_id", id)
		}
		return err
	},
})

func init() {
	// Command-specific flags.
	buildCmd.Flags().StringArrayVarP(&dirPaths, "dir", "d", []string{}, "0 or more directory paths")
	buildCmd.Flags().StringArrayVarP(&filePaths, "file", "f", []string{}, "0 or more file paths")
	kittehs.Must0(buildCmd.MarkFlagDirname("dir"))
	kittehs.Must0(buildCmd.MarkFlagFilename("file"))
	buildCmd.MarkFlagsOneRequired("dir", "file")
}
