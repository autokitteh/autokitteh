package systest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitToArgs(t *testing.T) {
	testCases := []struct {
		cmdArgs  string
		expected []string
	}{
		{
			cmdArgs:  `simple args1 x 1`,
			expected: []string{"simple", "args1", "x", "1"},
		},
		{
			// test [.:]
			cmdArgs:  `session start --build-id bld_03 --entrypoint main.star:main`,
			expected: []string{"session", "start", "--build-id", "bld_03", "--entrypoint", "main.star:main"},
		},
		{
			// test x=y, with int, dot, quotes
			cmdArgs:  `pass --input a=1 --input b=2.3 --input c="meow"`,
			expected: []string{"pass", "--input", "a=1", "--input", "b=2.3", "--input", `c="meow"`},
		},
		{
			// test schedule
			cmdArgs:  `--schedule1 "* * * * *" --schedule2 "0 0 */1 * *" --schedule3 "@every 1h3m2s"`,
			expected: []string{"--schedule1", "* * * * *", "--schedule2", "0 0 */1 * *", "--schedule3", "@every 1h3m2s"},
		},
		{
			// scv reader would return additional space as a field. Trim spaces
			cmdArgs:  ` aaa bb `,
			expected: []string{"aaa", "bb"},
		},
	}

	for _, tc := range testCases {
		result := splitToArgs(tc.cmdArgs)

		assert.Equal(t, len(tc.expected), len(result))
		assert.Equal(t, tc.expected, result)
	}
}
