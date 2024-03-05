package sdktypes

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

// Definitions here are applicable to all strs, ids and objects.

type (
	// Strict() makes a strict validation on the object using
	// its StrictValidate() method on its traits if available.
	stricter interface{ Strict() error }

	// IsValid() returns if the object contains a valid message.
	// (essentially, this could also be named IsNil, but following
	//  reflect's example and others, IsValid is more fitting)
	isValider interface{ IsValid() bool }
)

func ToString(o fmt.Stringer) string { return o.String() }

func Strict[T stricter](t T, err error) (T, error) {
	var zero T
	if err != nil {
		return zero, err
	}

	if err := t.Strict(); err != nil {
		return zero, err
	}

	return t, nil
}

func IsValid[V isValider](v V) bool { return v.IsValid() }

func hash[M comparableMessage](m M) string {
	var zero M
	if m == zero {
		return ""
	}

	hash := sha512.Sum512_256(kittehs.Must1(proto.Marshal(m)))
	return hex.EncodeToString(hash[:])
}
