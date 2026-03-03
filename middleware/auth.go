package middleware

import (
	"imp101/database"
	"imp101/models"
	"imp101/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware validates JWT token and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Handle missing role in token (backward compatibility)
		role := claims.Role
		if role == "" {
			// Fetch role from database using user_id
			var user models.User
			if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
					c.Abort()
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user role"})
				c.Abort()
				return
			}
			role = user.Role
			// Default to candidate if role is still empty
			if role == "" {
				role = models.RoleCandidate
			}
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", role)

		c.Next()
	}
}
