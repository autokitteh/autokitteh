package common

import (
	"regexp"
	"time"
)

// ParseGoTimestamp parses a [time.Time.String] string.
// It ignores unnecessary suffixes: sub-seconds and extra timezone details.
// Local example: "2025-02-28 10:04:21.024 -0800 PST m=+3759.281638293".
func ParseGoTimestamp(ts string) (time.Time, error) {
	// Get rid of the " PST m=..." suffix (keep the numeric timezone).
	ts = regexp.MustCompile(` [A-Z].*`).ReplaceAllString(ts, "")
	// Get rid of the timestamp's sub-second suffix
	// (it's both unnecessary and nondeterministic).
	ts = regexp.MustCompile(`\.\d+`).ReplaceAllString(ts, "")
	return time.Parse("2006-01-02 15:04:05 -0700", ts)
}
