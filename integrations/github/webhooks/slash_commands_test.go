package webhooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractSlashCommandsFromMD(t *testing.T) {
	tests := []struct {
		name     string
		md       string
		expected []slashCommand
	}{
		{
			name:     "empty string",
			md:       "",
			expected: nil,
		},
		{
			name:     "no slash commands",
			md:       "This is just regular markdown text\nwith multiple lines\nbut no commands",
			expected: nil,
		},
		{
			name: "single command without args",
			md:   "/deploy",
			expected: []slashCommand{
				{Name: "deploy", Args: []string{}, Raw: "/deploy"},
			},
		},
		{
			name: "single command with one arg",
			md:   "/deploy production",
			expected: []slashCommand{
				{Name: "deploy", Args: []string{"production"}, Raw: "/deploy production"},
			},
		},
		{
			name: "single command with multiple args",
			md:   "/deploy staging --force --verbose",
			expected: []slashCommand{
				{Name: "deploy", Args: []string{"staging", "--force", "--verbose"}, Raw: "/deploy staging --force --verbose"},
			},
		},
		{
			name: "multiple commands on separate lines",
			md:   "/build\n/test\n/deploy",
			expected: []slashCommand{
				{Name: "build", Args: []string{}, Raw: "/build"},
				{Name: "test", Args: []string{}, Raw: "/test"},
				{Name: "deploy", Args: []string{}, Raw: "/deploy"},
			},
		},
		{
			name: "commands mixed with regular text",
			md:   "Please review this PR\n/build\nLooks good to me\n/approve\nThanks!",
			expected: []slashCommand{
				{Name: "build", Args: []string{}, Raw: "/build"},
				{Name: "approve", Args: []string{}, Raw: "/approve"},
			},
		},
		{
			name: "command with trailing whitespace",
			md:   "/deploy production   \n/test  ",
			expected: []slashCommand{
				{Name: "deploy", Args: []string{"production"}, Raw: "/deploy production"},
				{Name: "test", Args: []string{}, Raw: "/test"},
			},
		},
		{
			name: "command with trailing tabs",
			md:   "/deploy\t\t\n/test arg\t",
			expected: []slashCommand{
				{Name: "deploy", Args: []string{}, Raw: "/deploy"},
				{Name: "test", Args: []string{"arg"}, Raw: "/test arg"},
			},
		},
		{
			name:     "slash not at beginning of line",
			md:       "Please run /deploy to deploy",
			expected: nil,
		},
		{
			name: "slash at beginning after newline",
			md:   "Please run:\n/deploy production",
			expected: []slashCommand{
				{Name: "deploy", Args: []string{"production"}, Raw: "/deploy production"},
			},
		},
		{
			name:     "only slash character",
			md:       "/",
			expected: nil,
		},
		{
			name:     "slash followed by only whitespace",
			md:       "/   \n/\t\t",
			expected: nil,
		},
		{
			name: "multiple args with various spacing",
			md:   "/run  arg1   arg2    arg3",
			expected: []slashCommand{
				{Name: "run", Args: []string{"arg1", "arg2", "arg3"}, Raw: "/run  arg1   arg2    arg3"},
			},
		},
		{
			name: "command with special characters in args",
			md:   "/notify @user #channel https://example.com",
			expected: []slashCommand{
				{Name: "notify", Args: []string{"@user", "#channel", "https://example.com"}, Raw: "/notify @user #channel https://example.com"},
			},
		},
		{
			name: "command names with numbers and underscores",
			md:   "/cmd_123\n/test2",
			expected: []slashCommand{
				{Name: "cmd_123", Args: []string{}, Raw: "/cmd_123"},
				{Name: "test2", Args: []string{}, Raw: "/test2"},
			},
		},
		{
			name: "windows-style line endings",
			md:   "/build\r\n/test\r\n/deploy",
			expected: []slashCommand{
				{Name: "build", Args: []string{}, Raw: "/build"},
				{Name: "test", Args: []string{}, Raw: "/test"},
				{Name: "deploy", Args: []string{}, Raw: "/deploy"},
			},
		},
		{
			name: "empty lines between commands",
			md:   "/start\n\n\n/middle\n\n/end",
			expected: []slashCommand{
				{Name: "start", Args: []string{}, Raw: "/start"},
				{Name: "middle", Args: []string{}, Raw: "/middle"},
				{Name: "end", Args: []string{}, Raw: "/end"},
			},
		},
		{
			name: "real world example - PR comment",
			md: `Thanks for the contribution!

/build all
/test integration

LGTM! Let's deploy this.

/approve
/deploy staging`,
			expected: []slashCommand{
				{Name: "build", Args: []string{"all"}, Raw: "/build all"},
				{Name: "test", Args: []string{"integration"}, Raw: "/test integration"},
				{Name: "approve", Args: []string{}, Raw: "/approve"},
				{Name: "deploy", Args: []string{"staging"}, Raw: "/deploy staging"},
			},
		},
		{
			name:     "command in code block should be ignored",
			md:       "```\n/deploy\n```",
			expected: nil,
		},
		{
			name:     "command in code block with language",
			md:       "```bash\n/deploy production\n```",
			expected: nil,
		},
		{
			name: "commands before and after code block",
			md:   "/before\n```\n/inside\n```\n/after",
			expected: []slashCommand{
				{Name: "before", Args: []string{}, Raw: "/before"},
				{Name: "after", Args: []string{}, Raw: "/after"},
			},
		},
		{
			name:     "command in tilde code block",
			md:       "~~~\n/deploy\n~~~",
			expected: nil,
		},
		{
			name: "nested code blocks not supported - toggles on each delimiter",
			md:   "```\n/first\n```\n/middle\n```\n/second\n```",
			expected: []slashCommand{
				{Name: "middle", Args: []string{}, Raw: "/middle"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSlashCommandsFromMD(tt.md)
			assert.Equal(t, tt.expected, result)
		})
	}
}
