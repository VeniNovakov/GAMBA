package transaction

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
	r.GET("/transactions", c.GetUserTransactions)
	r.GET("/transactions/summary", c.GetUserSummary)
	r.GET("/transactions/:id", c.GetByID)
	r.POST("/transactions/deposit", c.Deposit)
	r.POST("/transactions/withdraw", c.Withdraw)
	r.POST("/transactions/transfer", c.Transfer)
}

func (c *Controller) RegisterAdminRoutes(r *gin.RouterGroup) {
	r.GET("/admin/transactions", c.GetAll)
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user := auth.GetClaims(ctx)

	tx, err := c.service.GetByID(id, user.UserID, user.Role == models.RoleAdministrator)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, tx)
}

func (c *Controller) GetUserTransactions(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	var filter TransactionFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	transactions, err := c.service.GetUserTransactions(user.UserID, &filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, transactions)
}

func (c *Controller) GetAll(ctx *gin.Context) {

	var filter TransactionFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	transactions, err := c.service.GetAll(&filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, transactions)
}

func (c *Controller) Deposit(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	var req DepositRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tx, err := c.service.Deposit(user.UserID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, tx)
}

func (c *Controller) Withdraw(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	var req WithdrawRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tx, err := c.service.Withdraw(user.UserID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, tx)
}

func (c *Controller) Transfer(ctx *gin.Context) {
	user := auth.GetClaims(ctx)

	var req TransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tx, err := c.service.Transfer(user.UserID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, tx)
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
	case ErrTransactionNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrInsufficientFunds:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrInvalidAmount:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case ErrUserNotFound:
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrCannotTransferSelf:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
