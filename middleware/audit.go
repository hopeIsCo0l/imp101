package middleware

import (
	"encoding/json"

	"imp101/database"
	"imp101/models"

	"github.com/gin-gonic/gin"
)

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() < 400 && c.Request.Method == "GET" {
			return
		}

		var userID *uint
		if v, exists := c.Get("user_id"); exists {
			if id, ok := v.(uint); ok {
				userID = &id
			}
		}

		details, _ := json.Marshal(map[string]interface{}{
			"status": c.Writer.Status(),
		})

		_ = database.DB.Create(&models.AuditLog{
			UserID:    userID,
			Action:    "request",
			Endpoint:  c.FullPath(),
			Method:    c.Request.Method,
			IPAddress: c.ClientIP(),
			Details:   string(details),
		}).Error
	}
}
