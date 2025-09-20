package webhookssvc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestParseOutcomeValue(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		expected    httpOutcome
		expectError bool
	}{
		{
			name: "valid outcome with status and body",
			input: map[string]any{
				"status_code": 200,
				"body":        "hello world",
				"headers": map[string]any{
					"Content-Type": "text/plain",
				},
				"more": false,
			},
			expected: httpOutcome{
				StatusCode: 200,
				Body:       sdktypes.NewStringValue("hello world"),
				Headers:    map[string]string{"Content-Type": "text/plain"},
				More:       false,
			},
			expectError: false,
		},
		{
			name: "valid outcome with JSON",
			input: map[string]any{
				"status_code": 201,
				"json": map[string]any{
					"message": "created",
					"id":      42,
				},
				"headers": map[string]any{
					"Content-Type": "application/json",
				},
				"more": true,
			},
			expected: httpOutcome{
				StatusCode: 201,
				Headers:    map[string]string{"Content-Type": "application/json"},
				More:       true,
			},
			expectError: false,
		},
		{
			name: "minimal outcome",
			input: map[string]any{
				"status_code": 204,
			},
			expected: httpOutcome{
				StatusCode: 204,
			},
			expectError: false,
		},
		{
			name: "outcome with zero status",
			input: map[string]any{
				"status_code": 0,
				"body":        "default response",
			},
			expected: httpOutcome{
				StatusCode: 0,
				Body:       sdktypes.NewStringValue("default response"),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := sdktypes.WrapValue(tt.input)
			require.NoError(t, err, "failed to wrap input value")

			result, err := parseOutcomeValue(input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.StatusCode, result.StatusCode)
			assert.Equal(t, tt.expected.Headers, result.Headers)
			assert.Equal(t, tt.expected.More, result.More)

			if tt.expected.Body.IsValid() {
				assert.True(t, result.Body.IsValid())
				assert.Equal(t, tt.expected.Body.String(), result.Body.String())
			} else {
				assert.False(t, result.Body.IsValid())
			}

			if tt.expected.Json.IsValid() {
				assert.True(t, result.Json.IsValid())
				assert.Equal(t, tt.expected.Json.String(), result.Json.String())
			}
		})
	}
}

func TestHTTPOutcome_WriteBody(t *testing.T) {
	tests := []struct {
		name        string
		outcome     httpOutcome
		expectError bool
		validate    func(t *testing.T, buf *bytes.Buffer)
	}{
		{
			name: "string body",
			outcome: httpOutcome{
				Body: sdktypes.NewStringValue("hello world"),
			},
			expectError: false,
			validate: func(t *testing.T, buf *bytes.Buffer) {
				assert.Equal(t, []byte("hello world"), buf.Bytes())
			},
		},
		{
			name: "bytes body",
			outcome: httpOutcome{
				Body: sdktypes.NewBytesValue([]byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}),
			},
			expectError: false,
			validate: func(t *testing.T, buf *bytes.Buffer) {
				assert.Equal(t, []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}, buf.Bytes())
			},
		},
		{
			name: "JSON value in body field (marshaled)",
			outcome: httpOutcome{
				Body: func() sdktypes.Value {
					v, _ := sdktypes.WrapValue(map[string]any{
						"message": "hello",
						"count":   42,
					})
					return v
				}(),
			},
			expectError: false,
			validate: func(t *testing.T, buf *bytes.Buffer) {
				assert.NotEmpty(t, buf.Bytes())
				assert.Contains(t, buf.String(), "hello")
				assert.Contains(t, buf.String(), "42")
			},
		},
		{
			name: "JSON field",
			outcome: httpOutcome{
				Json: func() sdktypes.Value {
					v, _ := sdktypes.WrapValue(map[string]any{
						"status_code": "ok",
						"data":        []string{"a", "b", "c"},
					})
					return v
				}(),
			},
			expectError: false,
			validate: func(t *testing.T, buf *bytes.Buffer) {
				assert.NotEmpty(t, buf.Bytes())
				assert.Contains(t, buf.String(), "ok")
				assert.Contains(t, buf.String(), "a")
				assert.Contains(t, buf.String(), "b")
				assert.Contains(t, buf.String(), "c")
			},
		},
		{
			name: "both body and JSON set (error)",
			outcome: httpOutcome{
				Body: sdktypes.NewStringValue("body content"),
				Json: func() sdktypes.Value {
					v, _ := sdktypes.WrapValue(map[string]any{"json": "content"})
					return v
				}(),
			},
			expectError: true,
			validate:    nil,
		},
		{
			name:        "neither body nor JSON set",
			outcome:     httpOutcome{},
			expectError: false,
			validate: func(t *testing.T, buf *bytes.Buffer) {
				assert.Equal(t, "", buf.String())
			},
		},
		{
			name: "empty string body",
			outcome: httpOutcome{
				Body: sdktypes.NewStringValue(""),
			},
			expectError: false,
			validate: func(t *testing.T, buf *bytes.Buffer) {
				assert.Equal(t, "", buf.String())
			},
		},
		{
			name: "empty bytes body",
			outcome: httpOutcome{
				Body: sdktypes.NewBytesValue([]byte{}),
			},
			expectError: false,
			validate: func(t *testing.T, buf *bytes.Buffer) {
				assert.Equal(t, "", buf.String())
			},
		},
		{
			name: "complex nested JSON",
			outcome: httpOutcome{
				Json: func() sdktypes.Value {
					v, _ := sdktypes.WrapValue(map[string]any{
						"users": []any{
							map[string]any{"id": 1, "name": "Alice"},
							map[string]any{"id": 2, "name": "Bob"},
						},
						"meta": map[string]any{
							"total": 2,
							"page":  1,
						},
					})
					return v
				}(),
			},
			expectError: false,
			validate: func(t *testing.T, buf *bytes.Buffer) {
				assert.NotEmpty(t, buf.Bytes())
				resultStr := buf.String()
				assert.Contains(t, resultStr, "Alice")
				assert.Contains(t, resultStr, "Bob")
				assert.Contains(t, resultStr, "1")
				assert.Contains(t, resultStr, "2")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.outcome.WriteBody(&buf)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "outcome cannot have both 'body' and 'json' fields set together")
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, &buf)
			}
		})
	}
}

func TestHTTPOutcome_BodyBytes_ErrorCases(t *testing.T) {
	tests := []struct {
		name          string
		outcome       httpOutcome
		expectedError string
	}{
		{
			name: "both body and json set",
			outcome: httpOutcome{
				Body: sdktypes.NewStringValue("test"),
				Json: sdktypes.NewStringValue("test"),
			},
			expectedError: "outcome cannot have both 'body' and 'json' fields set together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := tt.outcome.WriteBody(&buf)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func TestHTTPOutcome_BodyBytes_Integration(t *testing.T) {
	inputValue := map[string]any{
		"status_code": 200,
		"body":        "Hello, World!",
		"headers": map[string]any{
			"Content-Type":   "text/plain",
			"Content-Length": "13",
		},
		"more": false,
	}

	wrappedValue, err := sdktypes.WrapValue(inputValue)
	require.NoError(t, err)

	outcome, err := parseOutcomeValue(wrappedValue)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = outcome.WriteBody(&buf)
	require.NoError(t, err)

	assert.Equal(t, []byte("Hello, World!"), buf.Bytes())
	assert.Equal(t, 200, outcome.StatusCode)
	assert.Equal(t, "text/plain", outcome.Headers["Content-Type"])
	assert.Equal(t, "13", outcome.Headers["Content-Length"])
	assert.False(t, outcome.More)
}
