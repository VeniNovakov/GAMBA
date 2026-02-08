package tournament

import (
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
	r.GET("/tournaments", c.GetAll)
	r.GET("/tournaments/:id", c.GetByID)
	r.GET("/tournaments/:id/leaderboard", c.GetLeaderboard)
	r.POST("/tournaments/:id/join", c.Join)
	r.POST("/tournaments/:id/leave", c.Leave)
}

func (c *Controller) RegisterAdminRoutes(r *gin.RouterGroup) {
	r.POST("/tournaments", c.Create)
	r.PUT("/tournaments/:id", c.Update)
	r.DELETE("/tournaments/:id", c.Delete)
	r.POST("/tournaments/:id/score", c.UpdateScore)
	r.POST("/tournaments/:id/end", c.EndTournament)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	var filter TournamentFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	tournaments, err := c.service.GetAll(&filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, tournaments)
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	tournament, err := c.service.GetByID(id)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, tournament)
}

func (c *Controller) Create(ctx *gin.Context) {

	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tournament, err := c.service.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusCreated, tournament)
}

func (c *Controller) Update(ctx *gin.Context) {

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

	tournament, err := c.service.Update(id, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, tournament)
}

func (c *Controller) Delete(ctx *gin.Context) {

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

func (c *Controller) Join(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	tournamentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	req := JoinRequest{TournamentID: tournamentID}
	participant, err := c.service.Join(user.UserID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, participant)
}

func (c *Controller) Leave(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	tournamentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.Leave(user.UserID, tournamentID); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "left tournament"})
}

func (c *Controller) UpdateScore(ctx *gin.Context) {

	tournamentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateScoreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := c.service.UpdateScore(tournamentID, &req); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "score updated"})
}

func (c *Controller) GetLeaderboard(ctx *gin.Context) {
	tournamentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	leaderboard, err := c.service.GetLeaderboard(tournamentID)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, leaderboard)
}

func (c *Controller) EndTournament(ctx *gin.Context) {

	tournamentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.EndTournament(tournamentID); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "tournament ended"})
}

func handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrTournamentNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrTournamentFull:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrTournamentNotOpen:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrAlreadyJoined:
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case ErrInsufficientFunds:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrNotParticipant:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrTournamentNotActive:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
