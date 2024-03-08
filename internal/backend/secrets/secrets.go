package secrets

import (
	"context"
	"fmt"
)

const (
	// basePath is the prefix for all secret names. "default" is redundant
	// but safe: to be able to enforce cross-environment isolation even if
	// multiple environments are using a shared secret management service.
	basePath = "autokitteh/default/integrations"
)

// Secrets is an internal, generic, minimalistic API for management of
// autokitteh user secrets. This interface in itself does not enforce
// isolation - its gRPC wrappers do (based on integration identity) -
// that's why this interface is internal and not meant for direct
// usage by autokitteh integrations.
type Secrets interface {
	// Set creates or replaces (i.e. overwrite, not update) a named secret of key-value
	// data. Data size limit = from 25 KiB to 1 MiB, depending on infrastructure:
	//   - https://developer.hashicorp.com/vault/docs/internals/limits
	//   - https://docs.aws.amazon.com/secretsmanager/latest/userguide/reference_limits.html
	//   - https://cloud.google.com/secret-manager/quotas
	//   - https://learn.microsoft.com/en-us/azure/key-vault/secrets/about-secrets
	Set(ctx context.Context, scope, name string, data map[string]string) error
	// Get retrieves the key-value data associated with a named secret.
	// If the name does not exist then we return nothing, not an error.
	Get(ctx context.Context, scope, name string) (map[string]string, error)
	// Append a token (as a key, with the current timestamp as the value)
	// to an existing secret, or create it if it doesn't exist already.
	Append(ctx context.Context, scope, name, token string) error
	// Delete permanently deletes all the metadata and versions of key-value
	// data of a named secret. Deleting a nonexistent name has no effect,
	// but isn't considered an error.
	Delete(ctx context.Context, scope, name string) error
}

func secretPath(prefix, name string) string {
	return fmt.Sprintf("%s/%s/%s", basePath, prefix, name)
}
