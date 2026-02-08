package event

import (
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
	// Public
	r.GET("/events", c.GetAll)
	r.GET("/events/:id", c.GetByID)

	// Auth required
	r.POST("/events/:id/bet", c.PlaceBet)
}

func (c *Controller) RegisterAdminRoutes(r *gin.RouterGroup) {
	r.POST("/events", c.Create)
	r.PUT("/events/:id", c.Update)
	r.DELETE("/events/:id", c.Delete)
	r.POST("/events/:id/outcomes", c.AddOutcome)
	r.PUT("/events/:id/outcomes/:outcomeId", c.UpdateOutcome)
	r.DELETE("/events/:id/outcomes/:outcomeId", c.DeleteOutcome)
	r.POST("/events/:id/settle", c.Settle)
	r.POST("/events/:id/cancel", c.Cancel)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	var filter EventFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	events, err := c.service.GetAll(&filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, events)
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	event, err := c.service.GetByID(id)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, event)
}

func (c *Controller) Create(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	event, err := c.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusCreated, event)
}

func (c *Controller) Update(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	event, err := c.service.Update(id, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, event)
}

func (c *Controller) Delete(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.Delete(id); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (c *Controller) AddOutcome(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req CreateOutcomeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	outcome, err := c.service.AddOutcome(eventID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, outcome)
}

func (c *Controller) UpdateOutcome(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	outcomeID, err := uuid.Parse(ctx.Param("outcomeId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid outcome id"})
		return
	}

	var req UpdateOutcomeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	outcome, err := c.service.UpdateOutcome(outcomeID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, outcome)
}

func (c *Controller) DeleteOutcome(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	outcomeID, err := uuid.Parse(ctx.Param("outcomeId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid outcome id"})
		return
	}

	if err := c.service.DeleteOutcome(outcomeID); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (c *Controller) PlaceBet(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req PlaceBetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	bet, err := c.service.PlaceBet(userID, eventID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, bet)
}

func (c *Controller) Settle(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req SettleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := c.service.Settle(eventID, &req); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "event settled"})
}

func (c *Controller) Cancel(ctx *gin.Context) {
	if !isAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	eventID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	if err := c.service.Cancel(eventID); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "event cancelled"})
}

func getUserID(ctx *gin.Context) uuid.UUID {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return uuid.Nil
	}
	uid, _ := userID.(uuid.UUID)
	return uid
}

func isAdmin(ctx *gin.Context) bool {
	role, _ := ctx.Get("role")
	return role == "administrator"
}

func handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrEventNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrOutcomeNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrEventNotBettable:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrEventAlreadySettled:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrInsufficientFunds:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrInvalidAmount:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
