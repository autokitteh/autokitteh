package kittehs

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
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
