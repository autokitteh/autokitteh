package kittehs

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"slices"
)

// Hash computes the SHA256 hash of the given input.
// It uses gob encoding to serialize the input before hashing.
func SHA256Hash(what any) (string, error) {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(what); err != nil {
		return "", fmt.Errorf("gob: %w", err)
	}

	sha := sha256.Sum256(b.Bytes())
	return hex.EncodeToString(sha[:]), nil
}

// Use this to get a stable hash of a map, as a map is not ordered.
func SHA256HashMap[K cmp.Ordered, V any](m map[K]V) (string, error) {
	type I struct {
		K K
		V V
	}

	var l []I

	for k, v := range m {
		l = append(l, I{K: k, V: v})
	}

	l = slices.SortedStableFunc(slices.Values(l), func(a, b I) int {
		return cmp.Compare(a.K, b.K)
	})

	return SHA256Hash(l)
}

// FNV1aHashString computes the FNV-1a hash of the given string.
func FNV1aHashString(what string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(what))
	return h.Sum64()
}
