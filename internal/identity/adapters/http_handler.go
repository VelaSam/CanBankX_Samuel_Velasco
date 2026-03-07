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
	router.POST("/login", h.Login)
	router.POST("/login/mfa", h.VerifyMFA)
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
		// But do I want to do this now?
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

func (h *IdentityHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.APIError{
			ErrorCode: "BAD_REQUEST",
			Message:   "Invalid request payload",
			Details:   []string{err.Error()},
		})
		return
	}

	res, apiErr := h.svc.Login(req)
	if apiErr != nil {
		status := http.StatusUnauthorized
		if apiErr.ErrorCode == "INTERNAL_ERROR" {
			status = http.StatusInternalServerError
		}
		c.JSON(status, apiErr)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *IdentityHandler) VerifyMFA(c *gin.Context) {
	var req service.VerifyMFARequest

	// if bind json failed
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apierrors.APIError{
			ErrorCode: "BAD_REQUEST",
			Message:   "Invalid request payload",
			Details:   []string{err.Error()},
		})
		return
	}

	res, apiErr := h.svc.VerifyMFA(req)
	// if verify mfa failed
	if apiErr != nil {
		status := http.StatusUnauthorized
		if apiErr.ErrorCode == "INTERNAL_ERROR" {
			status = http.StatusInternalServerError
		}
		c.JSON(status, apiErr)
		return
	}

	c.JSON(http.StatusOK, res)
}
