package systest

import (
	"bytes"
	"sync"
)

// mutexBuffer is a [bytes.Buffer], wrapped by a [sync.Mutex] to allow
// multiple goroutines to share it without data races. We redirect the AK
// server's output to it, to detect when the it's ready for the test to begin.
// We can't use channels because we don't control - and can't interrupt - the
// operation of [io.Copy] which uses this buffer as a destination.
type mutexBuffer struct {
	b *bytes.Buffer
	*sync.Mutex
}

// newMutexBuffer initializes a new [bytes.Buffer], wrapped by a [sync.Mutex].
func newMutexBuffer() *mutexBuffer {
	return &mutexBuffer{b: new(bytes.Buffer), Mutex: new(sync.Mutex)}
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If
// the buffer becomes too large, Write will panic with [ErrTooLarge].
func (m *mutexBuffer) Write(p []byte) (int, error) {
	m.Lock()
	defer m.Unlock()
	return m.b.Write(p)
}

// Bytes returns a slice of length b.Len() holding the unread portion of the
// buffer. The slice is valid for use only until the next buffer modification
// (that is, only until the next call to a method like [Buffer.Read], [Buffer.Write],
// [Buffer.Reset], or [Buffer.Truncate]). The slice aliases the buffer content
// at least until the next buffer modification, so immediate changes to the
// slice will affect the result of future reads.
func (m *mutexBuffer) Bytes() []byte {
	m.Lock()
	defer m.Unlock()
	return m.b.Bytes()
}

// String returns the contents of the unread portion of the buffer
// as a string. If the [Buffer] is a nil pointer, it returns "<nil>".
//
// To build strings more efficiently, see the strings.Builder type.
func (m *mutexBuffer) String() string {
	m.Lock()
	defer m.Unlock()
	return m.b.String()
}
