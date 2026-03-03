package middleware

import (
	"net/http"

	"imp101/models"

	"github.com/gin-gonic/gin"
)

func RecruiterOrAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == models.RoleRecruiter || models.IsAdminRole(role) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Recruiter or admin access required"})
		c.Abort()
	}
}
