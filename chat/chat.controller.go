package chat

import (
	"gamba/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Controller struct {
	service *Service
	hub     *Hub
}

func NewController(service *Service, hub *Hub) *Controller {
	return &Controller{service: service, hub: hub}
}

func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/chats", c.GetUserChats)
	r.POST("/chats", c.CreateChat)
	r.GET("/chats/:id", c.GetChatByID)
	r.GET("/chats/:id/messages", c.GetMessages)
	r.POST("/chats/:id/messages", c.SendMessage)
	r.POST("/chats/:id/read", c.MarkChatAsRead)
	r.POST("/messages/:id/read", c.MarkMessageAsRead)
}

func (c *Controller) RegisterWebSocket(rg *gin.RouterGroup) {
	rg.GET("", c.HandleWebSocket)
}

func (c *Controller) GetUserChats(ctx *gin.Context) {
	user := auth.GetClaims(ctx)
	if user.UserID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var filter ChatFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	chats, err := c.service.GetUserChats(user.UserID, &filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, chats)
}

func (c *Controller) GetChatByID(ctx *gin.Context) {
	user := auth.GetClaims(ctx)
	if user.UserID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	chatID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	chat, err := c.service.GetChatByID(chatID, user.UserID)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, chat)
}

func (c *Controller) CreateChat(ctx *gin.Context) {
	user := auth.GetClaims(ctx)
	if user.UserID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	chat, err := c.service.CreateChat(user.UserID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, chat)
}

func (c *Controller) GetMessages(ctx *gin.Context) {
	user := auth.GetClaims(ctx)
	if user.UserID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	chatID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	var filter MessageFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}

	messages, err := c.service.GetMessages(chatID, user.UserID, &filter)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, messages)
}

func (c *Controller) SendMessage(ctx *gin.Context) {
	user := auth.GetClaims(ctx)
	if user.UserID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	chatID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	var req SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	message, err := c.service.SendMessage(chatID, user.UserID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, message)
}

func (c *Controller) MarkChatAsRead(ctx *gin.Context) {
	user := auth.GetClaims(ctx)
	if user.UserID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	chatID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	if err := c.service.MarkChatAsRead(chatID, user.UserID); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

func (c *Controller) MarkMessageAsRead(ctx *gin.Context) {
	user := auth.GetClaims(ctx)
	if user.UserID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	messageID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}

	if err := c.service.MarkAsRead(messageID, user.UserID); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}

func (c *Controller) HandleWebSocket(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		ID:     uuid.New(),
		UserID: user.UserID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    c.hub,
	}

	c.hub.Register(client)

	go client.WritePump()
	go client.ReadPump(c.service)
}

func handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrChatNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrMessageNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrUserNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrCannotChatSelf:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrChatAlreadyExists:
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case ErrUnauthorized:
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
