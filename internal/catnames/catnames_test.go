package catnames

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPick(t *testing.T) {
	var i int
	pick := func(int) int { i++; return i }

	gen := NewGenerator(pick)

	expected := []string{
		"Affectionate Alice",
		"Agreeable Amber",
		"Amusing Amelia",
		"Beautiful Andreas",
		"Beloved Angel",
		"Big Angelina",
	}

	gens := make([]string, len(expected))
	for i := range expected {
		gens[i] = gen()
	}

	assert.Equal(t, expected, gens)
}
