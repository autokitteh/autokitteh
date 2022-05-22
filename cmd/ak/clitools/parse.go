package clitools

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
)

func ParseValuesArgs(args []string, wrapped bool) (l []*apivalues.Value, m map[string]*apivalues.Value, err error) {
	m = make(map[string]*apivalues.Value)

	for i, arg := range args {
		var (
			k string
			v *apivalues.Value
		)

		if k, v, err = ParseValueArg(arg, wrapped); err != nil {
			err = fmt.Errorf("arg %d: %w", i, err)
			return
		}

		if k == "" {
			l = append(l, v)
		} else {
			m[k] = v
		}
	}

	return
}

func ParseValueArg(text string, wrapped bool) (k string, v *apivalues.Value, err error) {
	parts := strings.SplitN(text, "=", 2)

	vtext := parts[0]

	if len(parts) == 2 {
		k = parts[0]
		vtext = parts[1]
	}

	if len(vtext) == 0 {
		err = fmt.Errorf("empty value")
		return
	}

	v = &apivalues.Value{}

	if wrapped {
		if err = v.UnmarshalJSON([]byte(vtext)); err != nil {
			v = nil
			err = fmt.Errorf("unmarshal: %w", err)
		}

		return
	}

	var m interface{}

	if err = yaml.Unmarshal([]byte(vtext), &m); err != nil {
		err = fmt.Errorf("unmarshal: %w", err)
		return
	}

	if v, err = apivalues.Wrap(m); err != nil {
		err = fmt.Errorf("unwrap: %w", err)
		return
	}

	return
}
