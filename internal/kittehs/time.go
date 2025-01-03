package kittehs

import "time"

var now = time.Now

func Now() time.Time { return now() }

// FreezeTime is used for testing.
func FreezeTimeForTest() {
	t := time.Now()
	now = func() time.Time { return t }
}
