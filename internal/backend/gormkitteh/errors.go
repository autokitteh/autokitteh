package gormkitteh

import (
	"errors"
)

var (
	ErrInvalidDSN  = errors.New("invalid DSN")
	ErrUnknownType = errors.New("unknown DSN type")
)
