package user

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
	r.GET("/users/me", c.GetProfile)
	r.PUT("/users/me", c.UpdateProfile)
	r.GET("/users/search", c.Search)
	r.GET("/users/:id", c.GetByID)
}

func (c *Controller) RegisterAdminRoutes(r *gin.RouterGroup) {
	r.GET("/users", c.GetAll)
	r.POST("/users/:id/restrict", c.Restrict)
	r.POST("/users/:id/unrestrict", c.Unrestrict)
	r.POST("/users/:id/deactivate", c.Deactivate)
	r.POST("/users/:id/activate", c.Activate)
}

func (c *Controller) GetProfile(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := c.service.GetProfile(userID)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (c *Controller) UpdateProfile(ctx *gin.Context) {
	userID := getUserID(ctx)
	if userID == uuid.Nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := c.service.UpdateProfile(userID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, user)
}
func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user, err := c.service.GetByID(id)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, user)
}

func (c *Controller) Search(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	users, err := c.service.Search(query, 10)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, users)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	var filter UserFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	users, err := c.service.GetAll(&filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, users)
}

func (c *Controller) Restrict(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.SetRestricted(id, true); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "user restricted"})
}

func (c *Controller) Unrestrict(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.SetRestricted(id, false); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "user unrestricted"})
}

func (c *Controller) Deactivate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.SetActive(id, false); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "user deactivated"})
}

func (c *Controller) Activate(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.SetActive(id, true); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "user activated"})
}

func getUserID(ctx *gin.Context) uuid.UUID {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return uuid.Nil
	}
	uid, _ := userID.(uuid.UUID)
	return uid
}

func handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrUserNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrUsernameExists:
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case ErrInvalidPassword:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
