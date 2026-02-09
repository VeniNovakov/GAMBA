data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./loader",
  ]
}

env "local" {
  src = data.external_schema.gorm.url
  dev = "postgres://postgres:password@localhost:5432/gamba_dev?sslmode=disable"
  url = "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
  migration {
    dir = "file://migrations"
  }
}

env "neon" {
  src = data.external_schema.gorm.url
  url = "postgresql://neondb_owner:npg_5Fo7czapIfCm@ep-purple-feather-a9dmt08d-pooler.gwc.azure.neon.tech/gamba_dev2?sslmode=require&channel_binding=require"

  migration {
    dir = "file://migrations"
  }
}