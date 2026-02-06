package main

import (
	"gamba/models"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
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
		&models.Tournament{},
		&models.TournamentParticipant{},
		&models.Bet{},
		&models.Ticket{},
		&models.TicketMessage{},
		&models.Transaction{},
	)
	if err != nil {
		panic(err)
	}
	io.WriteString(os.Stdout, stmts)
}
