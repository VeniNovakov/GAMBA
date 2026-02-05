package main

import (
	"GAMBA/internal/routes"
)

func main() {
	r := routes.SetupRouter()
	r.Run(":8080")
}
