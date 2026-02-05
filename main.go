package main

import (
	"GAMBA/internal/database"
)

func main() {
	cfg := database.NewConfigFromEnv()
	_ = database.MustConnect(cfg)
	defer database.Close()

	// Auto-migrate your models
	// database.AutoMigrate(db, &User{}, &Post{})

	println("Connected to database")

	// rest of your app...
}
