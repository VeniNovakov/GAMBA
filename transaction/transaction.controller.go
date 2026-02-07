package transaction

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
	r.GET("/transactions", c.GetUserTransactions)
	r.GET("/transactions/summary", c.GetUserSummary)
	r.GET("/transactions/:id", c.GetByID)
	r.POST("/transactions/deposit", c.Deposit)
	r.POST("/transactions/withdraw", c.Withdraw)
	r.POST("/transactions/transfer", c.Transfer)
	r.GET("/admin/transactions", c.GetAll) // admin only
}

func (c *Controller) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID, isAdmin := getUserFromContext(ctx)

	tx, err := c.service.GetByID(id, userID, isAdmin)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, tx)
}

func (c *Controller) GetUserTransactions(ctx *gin.Context) {
	userID, _ := getUserFromContext(ctx)

	var filter TransactionFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
		return
	}

	if filter.Limit == 0 {
		filter.Limit = 20
	}

	transactions, err := c.service.GetUserTransactions(userID, &filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	ctx.JSON(http.StatusOK, transactions)
}

func (c *Controller) GetAll(ctx *gin.Context) {
	_, isAdmin := getUserFromContext(ctx)
	if !isAdmin {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

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
	userID, _ := getUserFromContext(ctx)

	var req DepositRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tx, err := c.service.Deposit(userID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, tx)
}

func (c *Controller) Withdraw(ctx *gin.Context) {
	userID, _ := getUserFromContext(ctx)

	var req WithdrawRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tx, err := c.service.Withdraw(userID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, tx)
}

func (c *Controller) Transfer(ctx *gin.Context) {
	userID, _ := getUserFromContext(ctx)

	var req TransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	tx, err := c.service.Transfer(userID, &req)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, tx)
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
