package temporalite

type Config struct {
	Ephemeral        bool              `envconfig:"EPHEMERAL" json:"ephemeral" default:"true"`
	DBPath           string            `envconfig:"DB_PATH" json:"db_path"`
	Namespace        string            `envconfig:"NAMESPACE" json:"namespace" default:"default"`
	FrontendGRPCPort int               `envconfig:"FRONTEND_GRPC_PORT" json:"forntend_grpc_port" default:"7233"`
	UIPort           int               `envconfig:"UI_PORT" json:"ui_port" default:"8233"`
	UIEnabled        bool              `envconfig:"UI_ENABLED" json:"ui_enabled" default:"false"`
	SQLitePragmas    map[string]string `envconfig:"SQLITE_PRAGMAS" json:"sqlite_pragmas"`
}
