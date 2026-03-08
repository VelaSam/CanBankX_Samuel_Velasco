package adapters

import (
	"net/http"

	apierrors "canbankx/internal/pkg/errors"
	"canbankx/internal/transfers/service"

	"github.com/gin-gonic/gin"
)

type TransferHandler struct {
	svc service.TransferService
}

func NewTransferHandler(svc service.TransferService) *TransferHandler {
	return &TransferHandler{svc: svc}
}

func (h *TransferHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/", h.CreateTransfer)
}

func (h *TransferHandler) CreateTransfer(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, apierrors.ErrUnauthorized)
		return
	}

	var req service.ProcessTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.APIError{
			ErrorCode: "BAD_REQUEST",
			Message:   "Invalid request payload",
			Details:   []string{err.Error()},
		})
		return
	}

	transfer, apiErr := h.svc.ProcessTransfer(userID.(uint), req)
	if apiErr != nil {
		status := http.StatusInternalServerError
		if apiErr.ErrorCode == "BAD_REQUEST" {
			status = http.StatusBadRequest
		}
		c.JSON(status, apiErr)
		return
	}

	c.JSON(http.StatusCreated, transfer)
}
