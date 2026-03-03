package controllers

import (
	"imp101/database"
	"imp101/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetUser retrieves user information
func GetUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// If ID parameter is provided, check if it matches the authenticated user
	idParam := c.Param("id")
	if idParam != "" {
		id, err := strconv.ParseUint(idParam, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Only allow users to access their own data
		if uint(id) != userID.(uint) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Remove password from response
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// GetAllUsers retrieves all users (admin only)
func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var users []models.User
	if err := database.DB.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, users)
}

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func UpdateUserRole(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	valid := map[string]bool{
		models.RoleCandidate:     true,
		models.RoleRecruiter:     true,
		models.RoleAdministrator: true,
		models.RoleSuperAdmin:    true,
	}
	if !valid[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role value"})
		return
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("role", req.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": userID, "role": req.Role})
}

func UpdateUserStatus(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	valid := map[string]bool{
		models.UserStatusActive:   true,
		models.UserStatusInactive: true,
		models.UserStatusLocked:   true,
	}
	if !valid[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("status", req.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": userID, "status": req.Status})
}

// DeleteMyData anonymizes a candidate profile for privacy compliance.
func DeleteMyData(c *gin.Context) {
	userID := c.GetUint("user_id")
	anonEmail := "deleted_user_" + strconv.FormatUint(uint64(userID), 10) + "@example.invalid"
	anonPhone := "deleted_" + strconv.FormatUint(uint64(userID), 10)

	updates := map[string]interface{}{
		"email":             anonEmail,
		"phone":             anonPhone,
		"full_name":         "Deleted User",
		"nationality":       "",
		"date_of_birth":     nil,
		"status":            models.UserStatusInactive,
		"is_email_verified": false,
		"is_phone_verified": false,
		"updated_at":        time.Now(),
	}
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to anonymize account"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Account anonymized"})
}
