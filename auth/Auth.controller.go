package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *AuthService
}

func NewAuthController(authService *AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (c *AuthController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/register", c.Register)
	rg.POST("/login", c.Login)
	rg.POST("/refresh", c.Refresh)
	rg.POST("/logout", c.Logout)
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req registerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.authService.Register(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	tokens, err := c.authService.generateTokenPair(user)
	if err != nil {
		ctx.JSON(http.StatusCreated, gin.H{"message": "user created, please login manually"})
		return
	}

	ctx.JSON(http.StatusCreated, authResponse{
		User: userResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Role:     string(user.Role),
			Balance:  int64(user.Balance),
		},
		Tokens: tokens,
	})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := c.authService.Login(req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, ErrAccountInactive):
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, ErrAccountRestricted):
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		}
		return
	}

	ctx.JSON(http.StatusOK, authResponse{
		User: userResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Role:     string(user.Role),
			Balance:  int64(user.Balance),
		},
		Tokens: tokens,
	})
}

func (c *AuthController) Refresh(ctx *gin.Context) {
	var req refreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := c.authService.Refresh(req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken), errors.Is(err, ErrTokenRevoked):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, ErrAccountInactive):
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh token"})
		}
		return
	}

	ctx.JSON(http.StatusOK, tokens)
}

func (c *AuthController) Logout(ctx *gin.Context) {
	var req logoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.authService.Logout(req.RefreshToken); err != nil {
		if errors.Is(err, ErrInvalidToken) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
