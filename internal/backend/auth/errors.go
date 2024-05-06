package auth

import "errors"

var (
	ErrNotSupportedAuthProvider         = errors.New("not supported auth provider")
	ErrInvalidAuthProviderConfiguration = errors.New("invalid auth provider configuration")
)
