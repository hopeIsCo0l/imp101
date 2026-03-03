package models

import "time"

type ParsedCV struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ApplicationID uint      `json:"application_id" gorm:"uniqueIndex;not null"`
	ExtractedText string    `json:"extracted_text" gorm:"type:text"`
	TopSkillsJSON string    `json:"top_skills_json" gorm:"type:text"`
	CommonTerms   string    `json:"common_terms" gorm:"type:text"`
	Explanation   string    `json:"explanation" gorm:"type:text"`
	InitialScore  float64   `json:"initial_score"`
	ProcessedAt   time.Time `json:"processed_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
