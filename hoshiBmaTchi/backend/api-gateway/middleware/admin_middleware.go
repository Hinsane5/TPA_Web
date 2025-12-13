package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		role, exists := c.Get("role")
		
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Role verification failed"})
			c.Abort()
			return
		}

		if roleStr, ok := role.(string); !ok || roleStr != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}