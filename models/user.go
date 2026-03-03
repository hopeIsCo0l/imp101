package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleCandidate     = "candidate"
	RoleRecruiter     = "recruiter"
	RoleAdministrator = "administrator"
	RoleSuperAdmin    = "super_admin"
)

const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusLocked   = "locked"
)

type User struct {
	ID                  uint           `json:"id" gorm:"primaryKey"`
	Email               string         `json:"email" gorm:"uniqueIndex;not null"`
	Password            string         `json:"-" gorm:"not null"`
	FullName            string         `json:"full_name" gorm:"default:''"`
	Phone               string         `json:"phone" gorm:"index;default:''"`
	Nationality         string         `json:"nationality,omitempty"`
	DateOfBirth         *time.Time     `json:"date_of_birth,omitempty"`
	Role                string         `json:"role" gorm:"index;not null;default:'candidate'"`
	Status              string         `json:"status" gorm:"index;not null;default:'active'"`
	IsEmailVerified     bool           `json:"is_email_verified" gorm:"not null;default:false"`
	IsPhoneVerified     bool           `json:"is_phone_verified" gorm:"not null;default:false"`
	FailedLoginAttempts int            `json:"-" gorm:"not null;default:0"`
	LockedUntil         *time.Time     `json:"locked_until,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

type SignupRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	FullName    string `json:"full_name" binding:"required,min=3,max=120"`
	Phone       string `json:"phone" binding:"required,min=9,max=20"`
	Nationality string `json:"nationality"`
	DateOfBirth string `json:"date_of_birth"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

func IsAdminRole(role string) bool {
	return role == RoleAdministrator || role == RoleSuperAdmin
}
