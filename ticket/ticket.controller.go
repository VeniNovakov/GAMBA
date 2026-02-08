package ticket

import (
	"encoding/json"
	"gamba/auth"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("", c.Create)
}

func (c *Controller) RegisterAdminRoutes(r gin.IRoutes) {
	r.GET("", c.GetAll)
	r.GET("/:id", c.GetByID)
	r.PUT("/:id", c.Update)
	r.POST("/:id/close", c.Close)
	r.POST("/:id/messages", c.AddMessage)
}

func (c *Controller) GetAll(ctx *gin.Context) {

	user := auth.GetClaims(ctx)

	tickets, err := c.service.GetAll(user.UserID, user.Role)
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(ctx, http.StatusOK, tickets)
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid id")
		return
	}

	user := auth.GetClaims(ctx)

	ticket, err := c.service.GetByID(id, user.UserID, user.Role)
	if err != nil {
		handleError(ctx, err)
		return
	}

	writeJSON(ctx, http.StatusOK, ticket)
}

func (c *Controller) Create(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	var req CreateRequest
	if err := decodeJSONBody(ctx, &req); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid request")
		return
	}

	ticket, err := c.service.Create(user.UserID, &req)
	if err != nil {
		writeError(ctx, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(ctx, http.StatusCreated, ticket)
}

func (c *Controller) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid id")
		return
	}

	user := auth.GetClaims(ctx)

	var req UpdateRequest
	if err := decodeJSONBody(ctx, &req); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid request")
		return
	}

	ticket, err := c.service.Update(id, user.UserID, user.Role, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	writeJSON(ctx, http.StatusOK, ticket)
}

func (c *Controller) Close(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid id")
		return
	}

	user := auth.GetClaims(ctx)

	if err := c.service.Close(id, user.UserID); err != nil {
		handleError(ctx, err)
		return
	}

	writeJSON(ctx, http.StatusOK, gin.H{"message": "ticket closed"})
}

func (c *Controller) AddMessage(ctx *gin.Context) {
	ticketID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid id")
		return
	}

	user := auth.GetClaims(ctx)

	var req AddMessageRequest
	if err := decodeJSONBody(ctx, &req); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid request")
		return
	}

	message, err := c.service.AddMessage(ticketID, user.UserID, user.Role, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	writeJSON(ctx, http.StatusCreated, message)
}

func handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrTicketNotFound:
		writeError(ctx, http.StatusNotFound, err.Error())
	case ErrUnauthorized:
		writeError(ctx, http.StatusForbidden, err.Error())
	default:
		writeError(ctx, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(ctx *gin.Context, status int, data any) {
	ctx.JSON(status, data)
}

func writeError(ctx *gin.Context, status int, msg string) {
	ctx.JSON(status, gin.H{"error": msg})
}

func decodeJSONBody(ctx *gin.Context, dst any) error {
	dec := json.NewDecoder(ctx.Request.Body)
	dec.DisallowUnknownFields() // optional; remove if you want permissive parsing
	return dec.Decode(dst)
}
