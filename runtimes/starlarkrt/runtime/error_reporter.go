package runtime

import (
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type errorReporter struct {
	errs []string
}

func (er *errorReporter) Error(args ...any) { er.errs = append(er.errs, fmt.Sprint(args...)) }

func (er *errorReporter) Report() string {
	return strings.Join(
		kittehs.TransformWithIndex(er.errs, func(i int, err string) string {
			return fmt.Sprintf("FAIL %d: %s", i, err)
		}),
		"\n",
	)
}
