package sdktypes

import (
	"errors"
	"fmt"
	"testing"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func TestIsProgramError(t *testing.T) {
	var p programError

	if !errors.Is(p, sdkerrors.ErrProgram) {
		t.Errorf("expected true, got false")
	}

	if !errors.Is(fmt.Errorf("meow %w", p), sdkerrors.ErrProgram) {
		t.Errorf("expected true, got false")
	}
}
