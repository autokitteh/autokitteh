data "external_schema" "postgres" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/backend/db/dbgorm/scheme",
    "--dialect", "postgres",
  ]
}
data "external_schema" "postgres_enterprise" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/backend/db/dbgorm/scheme",
    "--dialect", "postgres",
    "--build-tags", "enterprise",
  ]
}

env "postgres" {
  src = data.external_schema.postgres.url
  dev = "docker://postgres/15/dev?search_path=public"
  migration {
    dir     = "file://migrations/postgres?format=goose"
    exclude = ["bases"]
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

# Environment for enterprise migrations
env "postgres_enterprise" {
  # Assume base has already been applied
  src = data.external_schema.postgres_enterprise.url
  dev = "docker://postgres/15/dev?search_path=public"
  migration {
    dir     = "file://migrations/postgres/enterprise?format=goose"
    exclude = ["bases"]
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}

data "external_schema" "sqlite" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/backend/db/dbgorm/scheme",
    "--dialect", "sqlite",
  ]
}

env "sqlite" {
  src = data.external_schema.sqlite.url
  dev = "sqlite://file?mode=memory&_fk=1"
  migration {
    dir = "file://migrations/sqlite?format=goose"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
