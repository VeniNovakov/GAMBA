package bet

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
	r.GET("/bets", c.GetUserBets)
	r.GET("/bets/summary", c.GetUserSummary)
	r.GET("/bets/:id", c.GetByID)
	r.GET("/admin/bets", c.GetAll) // admin only
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID, isAdmin := getUserFromContext(ctx)

	bet, err := c.service.GetByID(id, userID, isAdmin)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, bet)
}

func (c *Controller) GetUserBets(ctx *gin.Context) {
	userID, _ := getUserFromContext(ctx)

	var filter BetFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	bets, err := c.service.GetUserBets(userID, &filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, bets)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	_, isAdmin := getUserFromContext(ctx)
	if !isAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var filter BetFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	bets, err := c.service.GetAll(&filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, bets)
}

func (c *Controller) GetUserSummary(ctx *gin.Context) {
	userID, _ := getUserFromContext(ctx)

	summary, err := c.service.GetUserSummary(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, summary)
}

func getUserFromContext(ctx *gin.Context) (uuid.UUID, bool) {
	userID, _ := ctx.Get("user_id")
	role, _ := ctx.Get("role")
	uid, _ := userID.(uuid.UUID)
	return uid, role == "administrator"
}

func handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrBetNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
