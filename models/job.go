package models

import "time"

const (
	JobStatusDraft     = "draft"
	JobStatusPublished = "published"
	JobStatusClosed    = "closed"
	JobStatusArchived  = "archived"
)

type Job struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	Title           string     `json:"title" gorm:"not null;size:150;index"`
	Description     string     `json:"description" gorm:"not null;type:text"`
	RequiredSkills  string     `json:"required_skills" gorm:"type:text"`
	Qualifications  string     `json:"qualifications" gorm:"type:text"`
	CriteriaWeights string     `json:"criteria_weights" gorm:"type:text"`
	Deadline        *time.Time `json:"deadline,omitempty" gorm:"index"`
	Status          string     `json:"status" gorm:"not null;default:'draft';index"`
	CreatedBy       uint       `json:"created_by" gorm:"index;not null"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateJobRequest struct {
	Title           string `json:"title" binding:"required,min=5,max=150"`
	Description     string `json:"description" binding:"required,min=100,max=5000"`
	RequiredSkills  string `json:"required_skills" binding:"required"`
	Qualifications  string `json:"qualifications"`
	CriteriaWeights string `json:"criteria_weights"`
	Deadline        string `json:"deadline"`
}

type UpdateJobRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	RequiredSkills  string `json:"required_skills"`
	Qualifications  string `json:"qualifications"`
	CriteriaWeights string `json:"criteria_weights"`
	Deadline        string `json:"deadline"`
	Status          string `json:"status"`
}
