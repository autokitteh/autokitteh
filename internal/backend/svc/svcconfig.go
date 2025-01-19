package svc

import "go.autokitteh.dev/autokitteh/internal/backend/configset"

type svcConfig struct {
	RootRedirect    string `koanf:"root_redirect"`
	SeedObjectsPath string `koanf:"seed_objects_path"`
}

var svcConfigs = configset.Set[svcConfig]{
	Default: &svcConfig{
		RootRedirect: "https://autokitteh.com",
	},
	Dev: &svcConfig{
		RootRedirect: "/internal",
	},
}
