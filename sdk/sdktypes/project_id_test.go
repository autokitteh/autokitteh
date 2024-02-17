package sdktypes

import (
	"testing"
)

func TestProjectID(t *testing.T) {
	testID(t, idFuncs[ProjectID, projectIDTraits]{
		New:           NewProjectID,
		Parse:         ParseProjectID,
		StrictParse:   StrictParseProjectID,
		ParseIDOrName: ParseProjectIDOrName,
	})
}
