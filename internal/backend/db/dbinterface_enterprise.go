//go:build enterprise
// +build enterprise

package db

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type WorkflowExecutionRequest struct {
	SessionID sdktypes.SessionID
	Args      any
	Memo      map[string]string
}

type DB interface {
	Shared
}
