package runtime

import "go.autokitteh.dev/autokitteh/internal/backend/configset"

type Config struct{}

var Configs = configset.Set[Config]{
	Default: &Config{},
}
