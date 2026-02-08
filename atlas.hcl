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
  url = "postgresql://<user>:<password>@<host>/<db>?sslmode=require&channel_binding=require"
  migration {
    dir = "file://migrations"
  }
}
