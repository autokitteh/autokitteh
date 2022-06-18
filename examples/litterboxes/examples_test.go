package litterboxes

import "testing"

func TestDump(t *testing.T) {
	for k := range Examples {
		t.Log(k)
	}

	t.Log(JSONExamples)
}
