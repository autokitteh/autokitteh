package systest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
)

// TODO: parse http response.

func jq(data, query string) (string, error) {
	q, err := gojq.Parse(query)
	if err != nil {
		return "", fmt.Errorf("invalid jq query: %w", err)
	}

	var x any
	if err := json.Unmarshal([]byte(data), &x); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	iter := q.Run(x)

	var (
		vs []string
		ok = true
	)

	for ok {
		var v any
		if v, ok = iter.Next(); ok {
			vs = append(vs, fmt.Sprint(v))
		}
	}

	if len(vs) == 0 {
		return "", fmt.Errorf("jq query returned no results: %s", query)
	}

	return strings.Join(vs, ","), nil
}
