package sdktypes

import (
	"encoding/binary"
	"sync/atomic"

	"github.com/google/uuid"
)

var uuidGenerator = UUIDGenerator

func SetIDGenerator(f func() uuid.UUID) { uuidGenerator = f }

func NewUUID() uuid.UUID { return uuidGenerator() }

func UUIDGenerator() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

// To be used for testing only, when we expect a certain ID.
// First ID generated will be init+1.
func NewSequentialIDGeneratorForTesting(init uint64) func() uuid.UUID {
	var n atomic.Uint64
	n.Store(init)

	return func() uuid.UUID {
		n1 := n.Add(1)

		var b uuid.UUID
		binary.BigEndian.PutUint64(b[8:], n1)

		return b
	}
}

func intn(n int) int {
	uuid := NewUUID()
	ui64 := binary.BigEndian.Uint64(uuid[8:])
	return int(ui64 % uint64(n))
}
