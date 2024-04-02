package kittehs

import (
	"fmt"
	"hash/fnv"
	"net/url"
	"strings"
)

// This is silly, but string does not implement fmt.Stringer for some reaosn.
type String string

var _ fmt.Stringer = String("")

func (s String) String() string { return string(s) }

func ToString[T fmt.Stringer](t T) string { return t.String() }

// HashString32 computes the 32-bit FNV-1a hash of s in software.
// maphash.String does only 64bit version.
// (in Alan I trust: https://github.com/google/starlark-go/blob/f86470692795f8abcf9f837a3c53cf031c5a3d7e/starlark/hashtable.go#L435)
func HashString32(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func HashString64(s string) uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

func MatchLongestSuffix(s string, suffixes []string) string {
	var longest string
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			if len(longest) < len(suffix) {
				longest = suffix
			}
		}
	}
	return longest
}

func PadLeft(s string, r rune, n int) string {
	if len(s) >= n {
		return s
	}

	return strings.Repeat(string(r), n-len(s)) + s
}

// NormalizeURL ensures that the given URL has the right scheme
// prefix, and no suffix (e.g. path) after the host address.
func NormalizeURL(rawURL string, secure bool) (string, error) {
	// Normalize the URL's scheme prefix.
	scheme := "http://"
	if secure {
		scheme = "https://"
		rawURL = strings.TrimPrefix(rawURL, "http://")
	}
	if !strings.HasPrefix(rawURL, scheme) {
		rawURL = scheme + rawURL
	}

	// Parse the input URL.
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		return "", fmt.Errorf("no host in URL %q", rawURL)
	}

	// Reconstruct the URL with only the scheme and the host.
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}
