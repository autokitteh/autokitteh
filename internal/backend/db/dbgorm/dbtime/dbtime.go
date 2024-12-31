package dbtime

import "time"

var now = func() time.Time { return time.Now() }

func Now() time.Time { return now() }

// Used for testing.
func Freeze() {
	t := time.Now()
	now = func() time.Time { return t }
}
