package sdktypes

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
)

var idGenerator = UUIDGenerator

func SetIDGenerator(f func() string) { idGenerator = f }

func newID[T idTraits]() *id[T] {
	var t T

	return &id[T]{
		kind:  t.Kind(),
		value: idGenerator(),
	}
}

func UUIDGenerator() string {
	uuid := uuid.Must(uuid.NewUUID())
	return hex.EncodeToString(uuid[:])
}

// To be used for testing only, when we expect a certain ID.
// First ID generated will be init+1.
func NewSequentialIDGeneratorForTesting(init uint64) func() string {
	var n atomic.Uint64
	n.Store(init)

	return func() string {
		n1 := n.Add(1)

		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, n1)

		return fmt.Sprintf("%032s", hex.EncodeToString(b))
	}
}
