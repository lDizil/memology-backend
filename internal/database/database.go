package database

import (
	"fmt"
	"log"
	"time"

	"memology-backend/internal/config"
	"memology-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.DatabaseConfig) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	var db *gorm.DB
	var err error

	// Retry подключения к БД (до 30 секунд)
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Printf("Successfully connected to database on attempt %d", i+1)
			return db
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(3 * time.Second)
	}

	log.Fatal("Failed to connect to database after all retries:", err)
	return nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.UserSession{},
		&models.Meme{},
		&models.MemeMetrics{},
	)
}
