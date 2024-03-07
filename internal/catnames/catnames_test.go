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
		"Alexis The Agile",
		"Allie The Alert",
		"Ambra The Anxious",
		"Amethyst The Behavioral",
		"Andy The Best",
		"Angelica The Bossy",
	}

	var gens []string
	for range len(expected) {
		gens = append(gens, gen())
	}

	assert.Equal(t, expected, gens)
}
