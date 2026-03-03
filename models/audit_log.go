package models

import "time"

type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    *uint     `json:"user_id,omitempty" gorm:"index"`
	Action    string    `json:"action" gorm:"not null;index"`
	Endpoint  string    `json:"endpoint" gorm:"not null;index"`
	Method    string    `json:"method" gorm:"not null"`
	IPAddress string    `json:"ip_address" gorm:"not null"`
	Details   string    `json:"details" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}
