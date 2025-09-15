package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AIAuthMiddleware(apiToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-AI-Token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "AI token is required"})
			c.Abort()
			return
		}

		if token != apiToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid AI token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
