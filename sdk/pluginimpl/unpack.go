package pluginimpl

import (
	"fmt"
	"strings"

	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

func UnpackArgs(args []*apivalues.Value, kwargs_ map[string]*apivalues.Value, dsts ...interface{}) error {
	kwargs := make(map[string]*apivalues.Value, len(kwargs_))
	for k, v := range kwargs_ {
		kwargs[k] = v
	}

	if len(dsts)%2 != 0 {
		return fmt.Errorf("must have event number of dsts")
	}

	for i := 0; len(dsts) > 1; i++ {
		nameitf, dst := dsts[0], dsts[1]
		dsts = dsts[2:]

		name, ok := nameitf.(string)
		if !ok {
			return fmt.Errorf("dst %d name must be a string", i)
		}

		optional := strings.ContainsRune(name, '?')
		mustkw := strings.ContainsRune(name, '=')
		name = strings.TrimRight(name, "?=")

		v, found := kwargs[name]
		if found {
			delete(kwargs, name)
		} else {
			if len(args) > 0 && !mustkw {
				v, args = args[0], args[1:]
			} else {
				if !optional {
					return fmt.Errorf("required parameter %q not specified", name)
				}

				continue
			}
		}

		if err := apivalues.UnwrapInto(dst, v.Get()); err != nil {
			return fmt.Errorf("dst %q: %w", name, err)
		}
	}

	if len(args) > 0 {
		return fmt.Errorf("not all positional arguments consumed")
	}

	if len(kwargs) > 0 {
		return fmt.Errorf("not all keyword arguments consumed: %v", kwargs)
	}

	return nil
}
