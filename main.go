package main

import (
	"fmt"
	"gamba/auth"
	"gamba/chat"
	"gamba/ticket"
	"gamba/user"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func constructDsn() string {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, using system environment variables")
	}
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")
	channelBinding := os.Getenv("DB_CHANNELBINDING")

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=%s channelbinding=%s",
		host, user, password, dbname, sslmode, channelBinding,
	)
}

func main() {
	dsn := constructDsn()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	r := gin.Default()

	// Hub for WebSocket
	hub := chat.NewHub()
	go hub.Run()

	// Services
	authService := auth.NewAuthService(db, "your-jwt-secret")
	ticketService := ticket.NewService(db)
	chatService := chat.NewService(db, hub)
	userService := user.NewService(db)

	// Controllers
	authController := auth.NewAuthController(authService)
	ticketController := ticket.NewController(ticketService)
	chatController := chat.NewController(chatService, hub)
	userController := user.NewController(userService)

	// Public routes
	authRoutes := r.Group("/api/auth")
	authController.RegisterRoutes(authRoutes)

	// Protected routes
	api := r.Group("/api")
	api.Use(auth.Auth(authService))
	{
		ticketRoutes := api.Group("/tickets")
		ticketController.RegisterRoutes(ticketRoutes)

		chatController.RegisterRoutes(api)
		userController.RegisterRoutes(api)
	}

	chatController.RegisterWebSocket(r)

	// Admin routes
	admin := r.Group("/api/admin")
	admin.Use(auth.Auth(authService), auth.RequireRole("administrator"))
	{
		userController.RegisterAdminRoutes(admin)
	}

	r.Run(":8080")
}
