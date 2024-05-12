package authloginhttpsvc

import (
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

func matchLogin(p string) func(string) bool {
	p = strings.TrimSpace(p)

	matcher := func(login string) bool { return login == p }

	if p == "*" {
		// just "*" - match anything
		matcher = func(string) bool { return true }
	} else if username, host, _ := strings.Cut(p, "@"); username == "*" {
		// *@host - match any user at host
		matcher = func(login string) bool {
			_, host2, ok := strings.Cut(login, "@")
			return ok && host == host2
		}
	}

	return matcher
}

func compileLoginMatchers(patterns []string) func(string) bool {
	if len(patterns) == 0 {
		return func(string) bool { return true }
	}

	matchers := kittehs.Transform(patterns, matchLogin)

	return func(login string) bool {
		for _, m := range matchers {
			if m(login) {
				return true
			}
		}

		return false
	}
}
