package kittehs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	assert.Equal(t, ToString(String("meow")), "meow")
}

func TestPadLeft(t *testing.T) {
	assert.Equal(t, "01", PadLeft("1", '0', 2))
	assert.Equal(t, "1", PadLeft("1", '0', 1))
	assert.Equal(t, "000", PadLeft("", '0', 3))
	assert.Equal(t, "123", PadLeft("123", '0', 2))
}

func TestMatchLongetSuffix(t *testing.T) {
	assert.Equal(t, "", MatchLongestSuffix("", []string{"1", "3"}))
	assert.Equal(t, "234", MatchLongestSuffix("1234", []string{"4", "234", "34", "23"}))
}
