package ticket

import (
	"encoding/json"
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

// RegisterRoutes wires endpoints into a gin router/group.
func (c *Controller) RegisterRoutes(r gin.IRoutes) {
	r.GET("/tickets", c.GetAll)
	r.POST("/tickets", c.Create)
	r.GET("/tickets/:id", c.GetByID)
	r.PUT("/tickets/:id", c.Update)
	r.POST("/tickets/:id/close", c.Close)
	r.POST("/tickets/:id/messages", c.AddMessage)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	userID, isAdmin := getUserFromContext(ctx)

	tickets, err := c.service.GetAll(userID, isAdmin)
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

	userID, isAdmin := getUserFromContext(ctx)

	ticket, err := c.service.GetByID(id, userID, isAdmin)
	if err != nil {
		handleError(ctx, err)
		return
	}

	writeJSON(ctx, http.StatusOK, ticket)
}

func (c *Controller) Create(ctx *gin.Context) {
	userID, _ := getUserFromContext(ctx)

	var req CreateRequest
	if err := decodeJSONBody(ctx, &req); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid request")
		return
	}

	ticket, err := c.service.Create(userID, &req)
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

	userID, isAdmin := getUserFromContext(ctx)

	var req UpdateRequest
	if err := decodeJSONBody(ctx, &req); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid request")
		return
	}

	ticket, err := c.service.Update(id, userID, isAdmin, &req)
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

	userID, isAdmin := getUserFromContext(ctx)

	if err := c.service.Close(id, userID, isAdmin); err != nil {
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

	userID, isAdmin := getUserFromContext(ctx)

	var req AddMessageRequest
	if err := decodeJSONBody(ctx, &req); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid request")
		return
	}

	message, err := c.service.AddMessage(ticketID, userID, isAdmin, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}

	writeJSON(ctx, http.StatusCreated, message)
}

// getUserFromContext extracts user info from gin context (set by auth middleware).
// Your middleware should set these keys:
//
//	ctx.Set("user_id", uuid.UUID(...))
//	ctx.Set("role", "administrator")
func getUserFromContext(ctx *gin.Context) (uuid.UUID, bool) {
	var userID uuid.UUID
	if v, ok := ctx.Get("user_id"); ok {
		if id, ok := v.(uuid.UUID); ok {
			userID = id
		}
	}

	role := ""
	if v, ok := ctx.Get("role"); ok {
		if s, ok := v.(string); ok {
			role = s
		}
	}

	return userID, role == "administrator"
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

// decodeJSONBody mirrors your previous json.NewDecoder(r.Body).Decode(&req)
// and keeps behavior close to your chi version (no implicit validation).
func decodeJSONBody(ctx *gin.Context, dst any) error {
	dec := json.NewDecoder(ctx.Request.Body)
	dec.DisallowUnknownFields() // optional; remove if you want permissive parsing
	return dec.Decode(dst)
}
