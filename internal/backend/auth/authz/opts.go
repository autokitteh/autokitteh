package authz

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

type checkCfg struct {
	data                       map[string]any
	associations               map[string]sdktypes.ID
	convertForbiddenToNotFound bool
}

func configure(opts []CheckOpt) checkCfg {
	cfg := checkCfg{data: make(map[string]any) /* prevent nil */}
	for _, c := range opts {
		c(&cfg)
	}

	return cfg
}

type CheckOpt = func(*checkCfg)

func WithNop(*checkCfg) {}

// Set arbitrary data in check context.
func WithData(k string, v any) CheckOpt {
	return func(cfg *checkCfg) {
		if cfg.data == nil {
			cfg.data = make(map[string]any)
		}

		cfg.data[k] = v
	}
}

func WithFieldMask(fm *sdktypes.FieldMask) CheckOpt {
	return func(cfg *checkCfg) {
		if fm != nil {
			cfg.data["field_mask"] = fm.Paths
		}
	}
}

// Set `data.associated_<name>_org_id` in the check context.
// This will cause the checker to automatically deduce what org it belongs
// to based on the ID.
func WithAssociationWithID(name string, id sdktypes.ID) CheckOpt {
	return func(cfg *checkCfg) {
		if id == nil || !id.IsValid() {
			return
		}

		if cfg.associations == nil {
			cfg.associations = make(map[string]sdktypes.ID)
		}

		cfg.associations[name] = id
	}
}

func WithConvertForbiddenToNotFound(cfg *checkCfg) { cfg.convertForbiddenToNotFound = true }
