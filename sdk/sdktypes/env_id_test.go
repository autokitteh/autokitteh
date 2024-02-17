package sdktypes

import (
	"testing"
)

func TestEnvID(t *testing.T) {
	testID(t, idFuncs[EnvID, envIDTraits]{
		New:           NewEnvID,
		Parse:         ParseEnvID,
		StrictParse:   StrictParseEnvID,
		ParseIDOrName: ParseEnvIDOrName,
	})
}
