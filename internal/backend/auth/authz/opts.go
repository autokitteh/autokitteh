package authz

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type checkCfg struct {
	data                       map[string]any
	associations               map[string]sdktypes.ID
	convertForbiddenToNotFound bool
}

// Set arbitrary data in check context.
func WithData(k string, v any) func(*checkCfg) {
	return func(cfg *checkCfg) {
		if cfg.data == nil {
			cfg.data = make(map[string]any)
		}

		cfg.data[k] = v
	}
}

// Set `data.project_owner_id` in the check context.
// Useful to specify which project an object is created in.
// This will cause the checker to automatically deduce with project it belongs
// to based on the ID. For example: when creating a new event, the project will
// be deduced from the event destination id.
func WithAssociationWithID(name string, id sdktypes.ID) func(*checkCfg) {
	if id == nil || !id.IsValid() {
		return func(*checkCfg) {}
	}

	return func(cfg *checkCfg) {
		if cfg.associations == nil {
			cfg.associations = make(map[string]sdktypes.ID)
		}

		cfg.associations[name] = id
	}
}

func WithConvertForbiddenToNotFound(cfg *checkCfg) { cfg.convertForbiddenToNotFound = true }
