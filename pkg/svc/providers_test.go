//go:build unit

package svc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvidersSetGet(t *testing.T) {
	p := Providers{}

	p.Add("meow")
	p.Add(int(1))
	p.Add(1.2)
	p.Add(context.Background())

	var (
		i   int
		s   string
		f   float64
		ctx context.Context
	)

	if assert.True(t, p.Get(&s)) {
		assert.Equal(t, "meow", s)
	}

	if assert.True(t, p.Get(&f)) {
		assert.Equal(t, 1.2, f)
	}

	if assert.True(t, p.Get(&i)) {
		assert.Equal(t, 1, i)
	}

	if assert.True(t, p.Get(&ctx)) {
		assert.Equal(t, context.Background(), ctx)
	}

	assert.False(t, p.Get(&struct{}{}))
}
