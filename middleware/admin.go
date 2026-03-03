package middleware

import (
	"imp101/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware checks if the user has super_admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by AuthMiddleware)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		roleValue, ok := role.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role context"})
			c.Abort()
			return
		}

		// Check if user is administrator or super admin
		if !models.IsAdminRole(roleValue) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
