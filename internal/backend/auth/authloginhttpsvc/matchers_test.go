package authloginhttpsvc

import "testing"

func TestMatchLogin(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		logins  map[string]bool
	}{
		{
			name: "empty",
			logins: map[string]bool{
				"test":      false,
				"user@host": false,
				"user@evil": false,
				"user@":     false,
				"@host":     false,
			},
		},
		{
			name:    "any@host",
			pattern: "*@host",
			logins: map[string]bool{
				"test":      false,
				"user@host": true,
				"user@evil": false,
				"user@":     false,
				"@host":     true,
			},
		},
		{
			name:    "any",
			pattern: "*",
			logins: map[string]bool{
				"test":      true,
				"user@host": true,
				"user@evil": true,
				"user@":     true,
				"@host":     true,
			},
		},
		{
			name:    "exact",
			pattern: "user@host",
			logins: map[string]bool{
				"test":      false,
				"user@host": true,
				"user@evil": false,
				"user@":     false,
				"@host":     false,
			},
		},
		{
			name:    "wierd",
			pattern: "**@host",
			logins: map[string]bool{
				"test":      false,
				"user@host": false,
				"user@evil": false,
				"user@":     false,
				"@host":     false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			matcher := matchLogin(test.pattern)

			for login, expected := range test.logins {
				if result := matcher(login); result != expected {
					t.Errorf("unexpected %v for login %q", result, login)
				}
			}
		})
	}
}
