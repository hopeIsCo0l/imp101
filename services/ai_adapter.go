package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"imp101/database"
	"imp101/models"
)

type CVAnalysis struct {
	Score       float64
	TopSkills   []string
	CommonTerms []string
	Explanation string
	RawText     string
}

type AIAdapter interface {
	ProcessApplication(applicationID uint) error
}

type GoAIAdapter struct{}

func NewAIAdapter() AIAdapter {
	adapterMode := os.Getenv("AI_ADAPTER_MODE")
	if strings.EqualFold(adapterMode, "fastapi") {
		// Placeholder for future migration path.
		return &GoAIAdapter{}
	}
	return &GoAIAdapter{}
}

func (a *GoAIAdapter) ProcessApplication(applicationID uint) error {
	var app models.Application
	if err := database.DB.First(&app, applicationID).Error; err != nil {
		return err
	}

	var job models.Job
	if err := database.DB.First(&job, app.JobID).Error; err != nil {
		return err
	}

	cvBytes, err := os.ReadFile(app.CVFilePath)
	if err != nil {
		return err
	}

	cvText := string(cvBytes)
	analysis := ComputeTFIDFApproximation(cvText, job.Description, job.RequiredSkills)

	topSkillsJSON, _ := json.Marshal(analysis.TopSkills)
	parsed := models.ParsedCV{
		ApplicationID: app.ID,
		ExtractedText: analysis.RawText,
		TopSkillsJSON: string(topSkillsJSON),
		CommonTerms:   strings.Join(analysis.CommonTerms, ","),
		Explanation:   analysis.Explanation,
		InitialScore:  analysis.Score,
		ProcessedAt:   time.Now(),
	}

	if err := database.DB.Where("application_id = ?", app.ID).Assign(parsed).FirstOrCreate(&parsed).Error; err != nil {
		return err
	}

	app.CVScore = analysis.Score
	app.FinalScore = analysis.Score
	app.Status = models.ApplicationStatusParsed
	if err := database.DB.Save(&app).Error; err != nil {
		return err
	}

	return nil
}

func EnqueueApplicationProcessing(applicationID uint) {
	go func() {
		adapter := NewAIAdapter()
		if err := adapter.ProcessApplication(applicationID); err != nil {
			fmt.Printf("application processing failed: %v\n", err)
		}
	}()
}
