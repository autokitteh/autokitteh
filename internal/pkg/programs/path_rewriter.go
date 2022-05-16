package programs

import (
	"strings"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiprogram"
)

// If not rewritten, must return nil.
type PathRewriterFunc func(string) (_ *apiprogram.Path, name string, _ error)

// If in has prefix, return new path with newScheme and prefix dropped.
// Essentially map prefix into a scheme, and drop the prefix from the input path.
func NewPrefixPathRewriter(name, prefix, newScheme string) PathRewriterFunc {
	return func(in string) (*apiprogram.Path, string, error) {
		if !strings.HasPrefix(in, prefix) {
			return nil, "", nil
		}

		next, err := apiprogram.NewPath(newScheme, in[len(prefix):], "")

		return next, name, err
	}
}
