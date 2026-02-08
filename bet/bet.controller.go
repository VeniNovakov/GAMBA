package bet

import (
	"gamba/auth"
	"gamba/models"
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

}

func (c *Controller) RegisterAdminRoutes(r *gin.RouterGroup) {
	r.GET("/admin/bets", c.GetAll)
}
func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user := auth.GetClaims(ctx)

	bet, err := c.service.GetByID(id, user.UserID, user.Role == models.RoleAdministrator)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, bet)
}

func (c *Controller) GetUserBets(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	var filter BetFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	bets, err := c.service.GetUserBets(user.UserID, &filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, bets)
}

func (c *Controller) GetAll(ctx *gin.Context) {

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
	user := auth.GetClaims(ctx)

	summary, err := c.service.GetUserSummary(user.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, summary)
}

func handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrBetNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
