package models

import "time"

const (
	ApplicationStatusSubmitted = "submitted"
	ApplicationStatusParsed    = "parsed"
	ApplicationStatusRanked    = "ranked"
	ApplicationStatusRejected  = "rejected"
)

type Application struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	CandidateID    uint      `json:"candidate_id" gorm:"not null;index:idx_candidate_job,unique"`
	JobID          uint      `json:"job_id" gorm:"not null;index:idx_candidate_job,unique;index"`
	Status         string    `json:"status" gorm:"not null;default:'submitted';index"`
	CVFilePath     string    `json:"cv_file_path" gorm:"not null;type:text"`
	CoverLetter    string    `json:"cover_letter" gorm:"type:text"`
	CVScore        float64   `json:"cv_score" gorm:"default:0;index"`
	ExamScore      float64   `json:"exam_score" gorm:"default:0"`
	InterviewScore float64   `json:"interview_score" gorm:"default:0"`
	FinalScore     float64   `json:"final_score" gorm:"default:0;index"`
	SubmittedAt    time.Time `json:"submitted_at" gorm:"index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
