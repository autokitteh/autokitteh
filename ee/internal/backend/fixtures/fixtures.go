package fixtures

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	AutokittehOrgName           = kittehs.Must1(sdktypes.ParseSymbol("autokitteh"))
	AutokittehOrgID             = kittehs.Must1(sdktypes.ParseOrgID(fmt.Sprintf("o:%032x", 1)))
	AutokittehAnonymousUserName = "anonymous"
	AutokittehAnonymousUserID   = kittehs.Must1(sdktypes.ParseUserID(fmt.Sprintf("u:%032x", 0)))
)
