package fixtures

import "time"

var t0 = time.Now()

func Uptime() time.Duration { return time.Since(t0) }
