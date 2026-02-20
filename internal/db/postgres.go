package db

import (
	"log"
	"nutrition/internal/domain/user"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(databaseURL string) (*gorm.DB, error) {
	gormLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger,
	})
}

func Seed(database *gorm.DB) error {
	var count int64
	if err := database.Model(&user.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	testUser := user.User{
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
	}
	return database.Create(&testUser).Error
}
