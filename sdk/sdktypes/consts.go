package sdktypes

import "go.autokitteh.dev/autokitteh/internal/kittehs"

// REVIEW: we cannot define this in fixtures, due to circular dependencies
var BuiltinSchedulerConnectionID = kittehs.Must1(ParseConnectionID("con_3kthcr0n000000000000000000"))
