package manifest

import (
	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
)

var validateCmd = common.StandardCommand(&cobra.Command{
	Use:     "validate [file]",
	Short:   "Validate YAML manifest, from a file or stdin",
	Aliases: []string{"v"},
	Args:    cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		data, path, err := common.Consume(args)
		if err != nil {
			return err
		}

		if _, err := manifest.Read(data, path); err != nil {
			return err
		}

		return nil
	},
})
