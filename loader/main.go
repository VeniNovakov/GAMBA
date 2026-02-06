package main

import (
	"ariga.io/atlas-provider-gorm/gormschema"
	"gamba/models"
	"io"
	"os"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&models.User{},
		&models.Message{},
		&models.Chat{},
		&models.Game{},
		&models.Event{},
		&models.EventOutcome{},
		&models.RefreshToken{},
	)
	if err != nil {
		panic(err)
	}
	io.WriteString(os.Stdout, stmts)
}
