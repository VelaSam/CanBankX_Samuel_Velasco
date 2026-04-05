package adapters

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"canbankx/internal/pkg/bbc"
)

type BBCRequest struct {
	FromAccountID int     `json:"from_account_id" binding:"required"`
	ToKey         string  `json:"to_key" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
}

//RegisterBBCRoutes expose notes api routes externes pour bbc
func RegisterBBCRoutes(router *gin.RouterGroup) {
	router.POST("/transfers", InitiateBBCTransfer)
}

//InitiateBBCTransfer quand nos usagers veulent send du cash a une autre banque via le flow de bbc externe
func InitiateBBCTransfer(c *gin.Context) {
	var req BBCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}

	err := bbc.InitiateTransaction(req.FromAccountID, req.ToKey, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transfer forwarded to Central Bank (BBC) successfully.",
		"status":  "PENDING",
	})
}
