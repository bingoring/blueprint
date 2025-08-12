package database

import (
	"blueprint-module/pkg/config"
	"blueprint-module/pkg/database"
	localConfig "blueprint/internal/config"

	"gorm.io/gorm"
)

// GetDB returns the database connection
func GetDB() *gorm.DB {
	return database.GetDB()
}

// Connect initializes database connection using module
func Connect(cfg *localConfig.Config) error {
	// Convert local config to module config
	moduleConfig := &config.Config{
		Database: config.DatabaseConfig{
			Host:     cfg.Database.Host,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Name:     cfg.Database.Name,
			Port:     cfg.Database.Port,
			SSLMode:  cfg.Database.SSLMode,
		},
	}

	return database.Connect(moduleConfig)
}

// AutoMigrate runs database migrations
func AutoMigrate() error {
	return database.AutoMigrate()
}
