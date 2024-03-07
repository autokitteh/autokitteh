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
		"Alexis the Agile",
		"Allie the Alert",
		"Ambra the Anxious",
		"Amethyst the Behavioral",
		"Andy the Best",
		"Angelica the Bossy",
	}

	gens := make([]string, len(expected))
	for i := range expected {
		gens[i] = gen()
	}

	assert.Equal(t, expected, gens)
}
