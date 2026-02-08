package main

import (
	"fmt"
	"gamba/auth"
	"gamba/bet"
	"gamba/chat"
	"gamba/event"
	"gamba/game"
	"gamba/ticket"
	"gamba/tournament"
	"gamba/transaction"
	"gamba/user"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
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
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// Hub for WebSocket
	hub := chat.NewHub()
	go hub.Run()

	// Services
	authService := auth.NewAuthService(db, "your-jwt-secret")
	ticketService := ticket.NewService(db)
	chatService := chat.NewService(db, hub)
	userService := user.NewService(db)
	gameService := game.NewService(db)
	eventService := event.NewService(db)
	betService := bet.NewService(db)
	transactionService := transaction.NewService(db)
	tournamentsService := tournament.NewService(db)

	// Controllers
	authController := auth.NewAuthController(authService)
	ticketController := ticket.NewController(ticketService)
	chatController := chat.NewController(chatService, hub)
	userController := user.NewController(userService)
	gameController := game.NewController(gameService)
	eventController := event.NewController(eventService)
	betController := bet.NewController(betService)
	transactionController := transaction.NewController(transactionService)
	tournamentController := tournament.NewController(tournamentsService)

	// Public routes
	authRoutes := r.Group("/api/auth")
	authController.RegisterRoutes(authRoutes)

	// Protected routes
	api := r.Group("/api")
	//ticketRoutes := api.Group("/tickets")

	api.Use(auth.Auth(authService))
	{
		gameController.RegisterRoutes(api)
		chatController.RegisterRoutes(api)
		userController.RegisterRoutes(api)
		eventController.RegisterRoutes(api)
		betController.RegisterRoutes(api)
		transactionController.RegisterRoutes(api)
		tournamentController.RegisterRoutes(api)
		ticketController.RegisterRoutes(api)
	}

	ws := r.Group("/ws")
	ws.Use(auth.AuthWebsocket(authService))

	chatController.RegisterWebSocket(ws)

	// Admin routes
	admin := r.Group("/api")
	admin.Use(auth.Auth(authService), auth.RequireRole("administrator"))
	{
		ticketController.RegisterAdminRoutes(admin)
		gameController.RegisterAdminRoutes(admin)
		userController.RegisterAdminRoutes(admin)
		eventController.RegisterAdminRoutes(admin)
		betController.RegisterAdminRoutes(admin)
		transactionController.RegisterAdminRoutes(admin)
		tournamentController.RegisterAdminRoutes(admin)
	}

	r.Run(":8080")
}
