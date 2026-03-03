package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"imp101/database"
	"imp101/models"
	"imp101/services"

	"github.com/gin-gonic/gin"
)

func ApplyToJob(c *gin.Context) {
	candidateID := c.GetUint("user_id")
	role := c.GetString("role")
	if role != models.RoleCandidate {
		c.JSON(http.StatusForbidden, gin.H{"error": "Candidate role required"})
		return
	}

	jobID, err := strconv.Atoi(c.PostForm("job_id"))
	if err != nil || jobID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid job_id is required"})
		return
	}

	var job models.Job
	if err := database.DB.First(&job, jobID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	if job.Status != models.JobStatusPublished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job is not accepting applications"})
		return
	}
	if job.Deadline != nil && time.Now().After(*job.Deadline) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application deadline has passed"})
		return
	}

	var existing models.Application
	if err := database.DB.Where("candidate_id = ? AND job_id = ?", candidateID, jobID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You already applied to this job"})
		return
	}

	file, err := c.FormFile("cv")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CV file is required"})
		return
	}
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CV file must be <= 10MB"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".pdf": true, ".docx": true, ".png": true, ".jpg": true, ".jpeg": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Accepted: PDF, DOCX, PNG, JPG"})
		return
	}

	if err := os.MkdirAll("uploads/cv", 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare upload storage"})
		return
	}
	filename := strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + file.Filename
	savePath := filepath.Join("uploads", "cv", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store CV file"})
		return
	}

	app := models.Application{
		CandidateID: candidateID,
		JobID:       uint(jobID),
		Status:      models.ApplicationStatusSubmitted,
		CVFilePath:  savePath,
		CoverLetter: c.PostForm("cover_letter"),
		SubmittedAt: time.Now(),
	}
	if err := database.DB.Create(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create application"})
		return
	}

	services.EnqueueApplicationProcessing(app.ID)
	c.JSON(http.StatusAccepted, app)
}

func ListMyApplications(c *gin.Context) {
	candidateID := c.GetUint("user_id")
	var apps []models.Application
	if err := database.DB.Where("candidate_id = ?", candidateID).Order("created_at DESC").Find(&apps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}
	c.JSON(http.StatusOK, apps)
}

func GetMyApplication(c *gin.Context) {
	candidateID := c.GetUint("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}
	var app models.Application
	if err := database.DB.Where("id = ? AND candidate_id = ?", id, candidateID).First(&app).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Application not found"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func RankCandidatesByJob(c *gin.Context) {
	role := c.GetString("role")
	if role != models.RoleRecruiter && !models.IsAdminRole(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Recruiter or admin access required"})
		return
	}

	jobID, err := strconv.Atoi(c.Param("jobId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	var apps []models.Application
	if err := database.DB.Where("job_id = ?", jobID).Find(&apps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch applications"})
		return
	}

	sort.Slice(apps, func(i, j int) bool {
		return apps[i].FinalScore > apps[j].FinalScore
	})
	c.JSON(http.StatusOK, apps)
}

func GetApplicationExplainability(c *gin.Context) {
	role := c.GetString("role")
	if role != models.RoleRecruiter && !models.IsAdminRole(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Recruiter or admin access required"})
		return
	}
	appID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID"})
		return
	}

	var parsed models.ParsedCV
	if err := database.DB.Where("application_id = ?", appID).First(&parsed).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parsed analysis not found"})
		return
	}
	c.JSON(http.StatusOK, parsed)
}
