package controllers

import (
	"net/http"
	"strconv"
	"time"

	"imp101/database"
	"imp101/models"

	"github.com/gin-gonic/gin"
)

func parseDeadline(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func CreateJob(c *gin.Context) {
	role := c.GetString("role")
	if role != models.RoleRecruiter && !models.IsAdminRole(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Recruiter or admin access required"})
		return
	}

	var req models.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deadline, err := parseDeadline(req.Deadline)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deadline must use YYYY-MM-DD"})
		return
	}

	userID := c.GetUint("user_id")
	job := models.Job{
		Title:           req.Title,
		Description:     req.Description,
		RequiredSkills:  req.RequiredSkills,
		Qualifications:  req.Qualifications,
		CriteriaWeights: req.CriteriaWeights,
		Deadline:        deadline,
		Status:          models.JobStatusDraft,
		CreatedBy:       userID,
	}

	if err := database.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func ListJobs(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var jobs []models.Job
	query := database.DB.Model(&models.Job{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch jobs"})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func GetJob(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}
	var job models.Job
	if err := database.DB.First(&job, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func UpdateJob(c *gin.Context) {
	role := c.GetString("role")
	if role != models.RoleRecruiter && !models.IsAdminRole(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Recruiter or admin access required"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}
	var job models.Job
	if err := database.DB.First(&job, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	var req models.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.RequiredSkills != "" {
		updates["required_skills"] = req.RequiredSkills
	}
	if req.Qualifications != "" {
		updates["qualifications"] = req.Qualifications
	}
	if req.CriteriaWeights != "" {
		updates["criteria_weights"] = req.CriteriaWeights
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Deadline != "" {
		deadline, err := parseDeadline(req.Deadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "deadline must use YYYY-MM-DD"})
			return
		}
		updates["deadline"] = deadline
	}

	if err := database.DB.Model(&job).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job"})
		return
	}
	if err := database.DB.First(&job, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload job"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func PublishJob(c *gin.Context) {
	updateJobStatus(c, models.JobStatusPublished)
}

func CloseJob(c *gin.Context) {
	updateJobStatus(c, models.JobStatusClosed)
}

func ArchiveJob(c *gin.Context) {
	updateJobStatus(c, models.JobStatusArchived)
}

func updateJobStatus(c *gin.Context, status string) {
	role := c.GetString("role")
	if role != models.RoleRecruiter && !models.IsAdminRole(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Recruiter or admin access required"})
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}
	if err := database.DB.Model(&models.Job{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"job_id": id, "status": status})
}
