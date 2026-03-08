package adapters

import (
	"net/http"

	"canbankx/internal/accounts/service"
	apierrors "canbankx/internal/pkg/errors"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	svc service.AccountService
}

func NewAccountHandler(svc service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/", h.OpenAccount)
	router.GET("/", h.GetAccounts)
	router.GET("/:id/history", h.GetAccountHistory)
}

func (h *AccountHandler) OpenAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, apierrors.ErrUnauthorized)
		return
	}

	var req service.OpenAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.APIError{
			ErrorCode: "BAD_REQUEST",
			Message:   "Invalid request payload. Type must be CHECKING or SAVINGS.",
			Details:   []string{err.Error()},
		})
		return
	}

	account, apiErr := h.svc.OpenAccount(userID.(uint), req)
	if apiErr != nil {
		status := http.StatusInternalServerError
		if apiErr.ErrorCode == "BAD_REQUEST" {
			status = http.StatusBadRequest
		}
		c.JSON(status, apiErr)
		return
	}

	c.JSON(http.StatusCreated, account)
}

func (h *AccountHandler) GetAccounts(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, apierrors.ErrUnauthorized)
		return
	}

	accounts, apiErr := h.svc.GetAccounts(userID.(uint))
	if apiErr != nil {
		c.JSON(http.StatusInternalServerError, apiErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

func (h *AccountHandler) GetAccountHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, apierrors.ErrUnauthorized)
		return
	}

	var routeParams struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&routeParams); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.ErrInvalidInput)
		return
	}

	entries, balance, apiErr := h.svc.GetAccountHistory(userID.(uint), routeParams.ID)
	if apiErr != nil {
		status := http.StatusInternalServerError
		if apiErr.ErrorCode == "FORBIDDEN" {
			status = http.StatusForbidden
		}
		c.JSON(status, apiErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
		"entries": entries,
	})
}
