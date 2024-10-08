package projects

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/resolver"
)

var outputFile = "-"

var exportCmd = common.StandardCommand(&cobra.Command{
	Use:     "export <project name or ID>",
	Short:   "Export project",
	Aliases: []string{"ex"},
	Args:    cobra.ExactArgs(1),

	RunE: export,
})

var outputFileName string

func init() {
	exportCmd.Flags().StringVarP(&outputFileName, "output", "o", "-", "output file name (stdout by default)")
}

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

	zipData, err := r.Client.Projects().Export(ctx, prj.ID())
	if err != nil {
		return err
	}

	var out *os.File
	if outputFileName == "-" {
		out = os.Stdout
	} else {
		file, err := os.Create(outputFileName)
		if err != nil {
			return err
		}
		defer file.Close()
		out = file
	}

	if _, err := out.Write(zipData); err != nil {
		return err
	}

	return nil
}
