package sessions

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Common test time to use across all tests.
var fixedTime = time.Date(2025, 5, 15, 10, 30, 0, 0, time.UTC)

func TestWriteFormattedSessionLog_PrintRecord(t *testing.T) {
	buf := &bytes.Buffer{}

	// Create a print record with a string value.
	strValue := sdktypes.NewStringValue("Hello, world!")
	record := sdktypes.NewPrintSessionLogRecord(fixedTime, strValue, 1)

	err := writeFormattedSessionLog(buf, record)

	assert.NoError(t, err)
	assert.Equal(t, "[2025-05-15 10:30:00]:  Hello, world!\n", buf.String())
}

func TestWriteFormattedSessionLog_WithNewlines(t *testing.T) {
	buf := &bytes.Buffer{}

	// Create a print record with newlines.
	strValue := sdktypes.NewStringValue("First line\nSecond line\nThird line")
	record := sdktypes.NewPrintSessionLogRecord(fixedTime, strValue, 1)

	err := writeFormattedSessionLog(buf, record)

	assert.NoError(t, err)
	assert.Equal(t, "[2025-05-15 10:30:00]:  First line\n                        Second line\n                        Third line\n", buf.String())
}

func TestWriteFormattedSessionLog_WithEscapedNewlines(t *testing.T) {
	buf := &bytes.Buffer{}

	// Create a print record with escaped newlines.
	strValue := sdktypes.NewStringValue("First line\\nSecond line\\nThird line")
	record := sdktypes.NewPrintSessionLogRecord(fixedTime, strValue, 1)

	err := writeFormattedSessionLog(buf, record)

	assert.NoError(t, err)
	assert.Equal(t, "[2025-05-15 10:30:00]:  First line\n                        Second line\n                        Third line\n", buf.String())
}

func TestWriteFormattedSessionLog_EmptyPrint(t *testing.T) {
	buf := &bytes.Buffer{}

	// Create a print record with an empty string.
	strValue := sdktypes.NewStringValue("")
	record := sdktypes.NewPrintSessionLogRecord(fixedTime, strValue, 1)

	err := writeFormattedSessionLog(buf, record)

	assert.NoError(t, err)
	assert.Equal(t, "[2025-05-15 10:30:00]:  \n", buf.String())
}

func TestWriteFormattedSessionLog_WithEscapedQuotes(t *testing.T) {
	buf := &bytes.Buffer{}

	// Create a print record with escaped quotes.
	strValue := sdktypes.NewStringValue("He said: \\\"Hello\\\"")
	record := sdktypes.NewPrintSessionLogRecord(fixedTime, strValue, 1)

	err := writeFormattedSessionLog(buf, record)

	assert.NoError(t, err)
	assert.Equal(t, "[2025-05-15 10:30:00]:  He said: \"Hello\"\n", buf.String())
}

func TestWriteFormattedSessionLog_NonStringValue(t *testing.T) {
	buf := &bytes.Buffer{}

	// Create a print record with a non-string value (like a number).
	numValue := sdktypes.NewIntegerValue(42)
	record := sdktypes.NewPrintSessionLogRecord(fixedTime, numValue, 1)

	err := writeFormattedSessionLog(buf, record)

	assert.NoError(t, err)
	assert.Equal(t, "[2025-05-15 10:30:00]:  42\n", buf.String())
}

func TestWriteFormattedSessionLog_SkipInvalidRecord(t *testing.T) {
	buf := &bytes.Buffer{}

	// Use the invalid record
	record := sdktypes.InvalidSessionLogRecord

	err := writeFormattedSessionLog(buf, record)

	assert.NoError(t, err)
	assert.Equal(t, "", buf.String()) // Expecting no output
}

func TestWriteFormattedSessionLog_BufferWriteError(t *testing.T) {
	// Create a mock buffer that returns an error on write.
	mockBuf := &mockBuffer{}

	// Create a print record.
	strValue := sdktypes.NewStringValue("Test message")
	record := sdktypes.NewPrintSessionLogRecord(fixedTime, strValue, 1)

	err := writeFormattedSessionLog(mockBuf, record)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write log line")
}

// Mock buffer that returns an error on write.
type mockBuffer struct {
	bytes.Buffer
}

func (m *mockBuffer) WriteString(s string) (int, error) {
	return 0, assert.AnError
}
