package user

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
	r.GET("/users/me", c.GetProfile)
	r.PUT("/users/me", c.UpdateProfile)
	r.GET("/users/search", c.Search)
	r.GET("/users/:id", c.GetByID)

	r.GET("/friends", c.GetFriends)
	r.GET("/friends/requests", c.GetPendingRequests)
	r.GET("/friends/sent", c.GetSentRequests)
	r.POST("/friends/request", c.SendFriendRequest)
	r.POST("/friends/:id/accept", c.AcceptFriendRequest)
	r.POST("/friends/:id/reject", c.RejectFriendRequest)
	r.DELETE("/friends/:id", c.RemoveFriend)
}

func (c *Controller) RegisterAdminRoutes(r *gin.RouterGroup) {
	r.GET("/users", c.GetAll)
	r.POST("/users/:id/restrict", c.Restrict)
	r.POST("/users/:id/unrestrict", c.Unrestrict)
	r.POST("/users/:id/deactivate", c.Deactivate)
	r.POST("/users/:id/activate", c.Activate)
}

func (c *Controller) GetProfile(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	userEntity, err := c.service.GetProfile(user.UserID)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, userEntity)
}

func (c *Controller) UpdateProfile(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	var req UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userEntity, err := c.service.UpdateProfile(user.UserID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, userEntity)
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

func handleError(ctx *gin.Context, err error) {
	switch err {
	case ErrUserNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrUsernameExists:
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case ErrInvalidPassword:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrCannotFriendSelf:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrAlreadyFriends:
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case ErrRequestNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrNotFriends:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func (c *Controller) GetFriends(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	friends, err := c.service.GetFriends(user.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, friends)
}

func (c *Controller) GetPendingRequests(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	requests, err := c.service.GetPendingRequests(user.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, requests)
}

func (c *Controller) GetSentRequests(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	requests, err := c.service.GetSentRequests(user.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, requests)
}

func (c *Controller) SendFriendRequest(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	var req SendFriendRequestRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	friendID, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	request, err := c.service.SendFriendRequest(user.UserID, friendID)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, request)
}

func (c *Controller) AcceptFriendRequest(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	requestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.AcceptFriendRequest(user.UserID, requestID); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "friend request accepted"})
}

func (c *Controller) RejectFriendRequest(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	requestID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.RejectFriendRequest(user.UserID, requestID); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "friend request rejected"})
}

func (c *Controller) RemoveFriend(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	friendID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.service.RemoveFriend(user.UserID, friendID); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "friend removed"})
}
