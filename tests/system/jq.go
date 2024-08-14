package systest

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/itchyny/gojq"
)

var captures = make(map[string]string)

// TODO: parse http response.

func captureJQ(t *testing.T, step string, ak *akResult, _ *httpResponse) error {
	match := jqCheck.FindStringSubmatch(step)
	name, query := match[1], match[2]

	q, err := gojq.Parse(query)
	if err != nil {
		return fmt.Errorf("invalid jq query: %w", err)
	}

	var x any
	if err := json.Unmarshal([]byte(ak.output), &x); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
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
		return fmt.Errorf("jq query returned no results: %s", query)
	}

	t.Logf("captured %q into %q", vs, name)

	captures[name] = strings.Join(vs, ",")

	return nil
}

func expandCapture(s string) string {
	return os.Expand(s, func(key string) string {
		return captures[key]
	})
}
