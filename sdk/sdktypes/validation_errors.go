package sdktypes

import (
	"errors"
)

var errMissingFields = errors.New("missing fields")

func ensureNotEmpty(vs ...string) error { // TODO: map? to return which field is missing
	for _, v := range vs {
		if v == "" {
			return errMissingFields
		}
	}

	return nil
}
