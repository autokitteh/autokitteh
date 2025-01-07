package kittehs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/apipb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestProtoToMap(t *testing.T) {
	pb := &apipb.Method{
		Name:              "meow",
		RequestTypeUrl:    "woof",
		ResponseStreaming: true,
		Syntax:            1,
	}

	m, err := ProtoToMap(pb, Must1(fieldmaskpb.New(pb, "name", "response_streaming", "syntax")))
	if assert.NoError(t, err) {
		assert.Equal(t, map[string]interface{}{
			"name":               "meow",
			"response_streaming": true,
			"syntax":             1,
		}, m)
	}

	m, err = ProtoToMap(pb, nil)
	if assert.NoError(t, err) {
		assert.Len(t, m, 7)
	}
}
