package adapters

import (
	"net/http"

	"canbankx/internal/identity/service"
	apierrors "canbankx/internal/pkg/errors"

	"github.com/gin-gonic/gin"
)

type IdentityHandler struct {
	svc service.IdentityService
}

func NewIdentityHandler(svc service.IdentityService) *IdentityHandler {
	return &IdentityHandler{svc: svc}
}

func (h *IdentityHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/register", h.RegisterUser)
}

func (h *IdentityHandler) RegisterUser(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.APIError{
			ErrorCode: "BAD_REQUEST",
			Message:   "Invalid request payload",
			Details:   []string{err.Error()},
		})
		return
	}

	user, apiErr := h.svc.RegisterUser(req)
	if apiErr != nil {
		// A full robust implementation would map the ErrCode to specific HTTP Statuses (ADR-0003)
		status := http.StatusBadRequest
		if apiErr.ErrorCode == "INTERNAL_ERROR" {
			status = http.StatusInternalServerError
		}

		c.JSON(status, apiErr)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully in PENDING state",
		"userId":  user.ID,
		"status":  user.Status,
	})
}
