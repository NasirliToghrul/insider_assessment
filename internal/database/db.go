package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"case_study/internal/config"
	"case_study/internal/models"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	switch cfg.DBDriver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER: %s", cfg.DBDriver)
	}
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Message{}); err != nil {
		return nil, err
	}
	log.Println("database connected and migrated")
	return db, nil
}
