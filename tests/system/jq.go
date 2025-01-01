package systest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bcicen/jstream"
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
		d := jstream.NewDecoder(bytes.NewReader([]byte(data)), 0)

		var xs []any

		for v := range d.Stream() {
			if d.Err() != nil {
				break
			}

			xs = append(xs, v.Value)
		}

		if err := d.Err(); err != nil {
			return "", fmt.Errorf("invalid JSON: %w", err)
		}

		x = xs
	}

	iter := q.Run(x)

	var (
		vs []string
		ok = true
	)

	for ok {
		var v any
		if v, ok = iter.Next(); ok {
			if err, ok := v.(error); ok {
				return "", fmt.Errorf("jq query failed: %w", err)
			}

			vs = append(vs, fmt.Sprint(v))
		}
	}

	if len(vs) == 0 {
		return "", fmt.Errorf("jq query returned no results: %s", query)
	}

	return strings.Join(vs, ","), nil
}
