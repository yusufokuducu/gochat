package database

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/faust-lvii/gochat/backend/internal/models"
)

// DB is the database connection
var DB *gorm.DB

// Initialize initializes the database connection
func Initialize(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Warning: Failed to connect to database:", err)
		log.Println("Running in memory mode without database")
		return err
	}

	// Auto migrate the schema
	err = DB.AutoMigrate(&models.User{}, &models.Friendship{}, &models.Message{})
	if err != nil {
		log.Println("Failed to migrate database schema:", err)
		return err
	}

	// Create admin user if not exists
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		// Hash password: admin123
		passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Println("Failed to hash admin password:", err)
			return err
		}
		
		adminUser := models.User{
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: string(passwordHash),
			CreatedAt:    time.Now(),
		}
		
		if err := DB.Create(&adminUser).Error; err != nil {
			log.Println("Failed to create admin user:", err)
			return err
		}
		
		log.Println("Created admin user")
	}

	return nil
}

// GetDB returns the database connection
func GetDB() *gorm.DB {
	return DB
}
