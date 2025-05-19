package migrations

import "embed"

//go:embed **/*.sql
//go:embed **/**/*.sql
var Migrations embed.FS
