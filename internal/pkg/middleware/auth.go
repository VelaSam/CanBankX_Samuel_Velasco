package middleware

import (
	"net/http"
	"strings"

	"canbankx/internal/pkg/auth"
	apierrors "canbankx/internal/pkg/errors"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apierrors.ErrUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apierrors.ErrUnauthorized)
			return
		}

		tokenStr := parts[1]
		claims, err := auth.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apierrors.ErrUnauthorized)
			return
		}

		// ABSOLUTELY require USER role for these protected routes
		if claims.Role != "USER" {
			c.AbortWithStatusJSON(http.StatusForbidden, &apierrors.APIError{
				ErrorCode: "FORBIDDEN",
				Message:   "Insufficient permissions: requires FULL access",
			})
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}
