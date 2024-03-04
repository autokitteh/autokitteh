package manifest

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/cmd/ak/common"
	"go.autokitteh.dev/autokitteh/internal/manifest"
)

var execCmd = common.StandardCommand(&cobra.Command{
	Use:     "execute [file]",
	Short:   "Execute plan output from file or stdin",
	Aliases: []string{"exec", "x"},
	Args:    cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := common.Consume(args)
		if err != nil {
			return err
		}

		var actions manifest.Actions

		if err := json.Unmarshal(data, &actions); err != nil {
			return fmt.Errorf("invalid plan input: %w", err)
		}

		ctx, cancel := common.LimitedContext()
		defer cancel()

		_, err = manifest.Execute(ctx, actions, common.Client(), func(msg string) {
			fmt.Fprintf(os.Stderr, "[exec] %s\n", msg)
		})
		return err
	},
})
