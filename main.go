package main

import (
	"gamba/auth"
	"gamba/ticket"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=gamba_dev port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	r := gin.Default()

	authService := auth.NewAuthService(db, "your-jwt-secret")
	authController := auth.NewAuthController(authService)
	ticketService := ticket.NewService(db)
	ticketController := ticket.NewController(ticketService)
	// Public routes
	authRoutes := r.Group("/api/auth")
	authController.RegisterRoutes(authRoutes)

	// Protected routes
	api := r.Group("/api")

	api.Use(auth.Auth(authService))
	{
		ticketRoutes := api.Group("/tickets")
		ticketController.RegisterRoutes(ticketRoutes)
	}

	api.Use(auth.Auth(authService))
	admin := r.Group("/api/admin")
	admin.Use(auth.Auth(authService), auth.RequireRole("administrator"))
	{

	}

	r.Run(":8080")
}
