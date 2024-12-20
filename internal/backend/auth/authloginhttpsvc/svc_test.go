package authloginhttpsvc

import (
	"context"
	"testing"
)

func TestSuccessLoginHandler(t *testing.T) {
	a := &svc{}

	ctx := context.Background()

	_ = a.newSuccessLoginHandler(ctx, &loginData{})

	// TODO
}
