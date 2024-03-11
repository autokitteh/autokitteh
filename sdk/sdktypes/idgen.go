package sdktypes

import (
	"encoding/binary"
	"sync/atomic"

	"github.com/google/uuid"
)

var uuidGenerator = UUIDGenerator

type UUID = uuid.UUID

func SetIDGenerator(f func() UUID) { uuidGenerator = f }

func newUUID() []byte { id := uuidGenerator(); return id[:] }

func UUIDGenerator() UUID { return uuid.Must(uuid.NewUUID()) }

// To be used for testing only, when we expect a certain ID.
// First ID generated will be init+1.
func NewSequentialIDGeneratorForTesting(init uint64) func() UUID {
	var n atomic.Uint64
	n.Store(init)

	return func() UUID {
		n1 := n.Add(1)

		var b UUID
		binary.BigEndian.PutUint64(b[8:], n1)

		return b
	}
}

func intn(n int) int {
	uuid := newUUID()
	ui64 := binary.BigEndian.Uint64(uuid[8:])
	return int(ui64 % uint64(n))
}
