package common

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

func AddFailIfNotFoundFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("fail", "f", false, "fail if not found")
}

func FailIfNotFound[T any](cmd *cobra.Command, what string, v *T) error {
	if kittehs.Must1(cmd.Flags().GetBool("fail")) && v == nil {
		return NewExitCodeError(NotFoundExitCode, fmt.Errorf("%s not found", what))
	}
	return nil
}
