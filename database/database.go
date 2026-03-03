package database

import (
	"fmt"
	"imp101/models"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the database connection
func InitDB() error {
	// Get database connection details from environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "imp101"
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	// Construct PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// Auto migrate the schema
	err = DB.AutoMigrate(
		&models.User{},
		&models.Job{},
		&models.Application{},
		&models.ParsedCV{},
		&models.AuditLog{},
	)
	if err != nil {
		return err
	}

	// Seed super admin user
	err = seedSuperAdmin()
	if err != nil {
		return err
	}

	return nil
}

// seedSuperAdmin creates a super admin user if it doesn't exist, or updates it if it exists but doesn't have the correct role/password
func seedSuperAdmin() error {
	const adminEmail = "admin@admin.admin"
	const adminPassword = "CqZP99nfbUI2M#3"
	const adminRole = models.RoleSuperAdmin

	var existingUser models.User
	result := DB.Where("email = ?", adminEmail).First(&existingUser)

	if result.Error == nil {
		// User exists - check if role and password need to be updated
		needsUpdate := false
		updates := make(map[string]interface{})

		// Check if role is correct
		if existingUser.Role != adminRole {
			updates["role"] = adminRole
			needsUpdate = true
		}
		if existingUser.Status != models.UserStatusActive {
			updates["status"] = models.UserStatusActive
			needsUpdate = true
		}
		if existingUser.FullName == "" {
			updates["full_name"] = "System Super Admin"
			needsUpdate = true
		}
		if existingUser.Phone == "" {
			updates["phone"] = "0000000000"
			needsUpdate = true
		}

		// Check if password is correct
		err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(adminPassword))
		if err != nil {
			// Password doesn't match, update it
			hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			if hashErr != nil {
				return fmt.Errorf("failed to hash super admin password: %w", hashErr)
			}
			updates["password"] = string(hashedPassword)
			needsUpdate = true
		}

		// Update user if needed
		if needsUpdate {
			if err := DB.Model(&existingUser).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update super admin user: %w", err)
			}
		}

		return nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		// Database error occurred
		return result.Error
	}

	// User doesn't exist - create it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash super admin password: %w", err)
	}

	superAdmin := models.User{
		Email:           adminEmail,
		Password:        string(hashedPassword),
		FullName:        "System Super Admin",
		Phone:           "0000000000",
		Role:            adminRole,
		Status:          models.UserStatusActive,
		IsEmailVerified: true,
		IsPhoneVerified: true,
	}

	if err := DB.Create(&superAdmin).Error; err != nil {
		return fmt.Errorf("failed to create super admin user: %w", err)
	}

	return nil
}
