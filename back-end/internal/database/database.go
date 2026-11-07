package database

import (
	"fmt"
	"log"
	"time"

	"github.com/putita-lek/goodbuy-backend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a MySQL connection, retrying while the database starts up.
func Connect(dsn string) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	for attempt := 1; attempt <= 15; attempt++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
		if err == nil {
			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				pingErr = sqlDB.Ping()
			}
			if pingErr == nil {
				sqlDB.SetMaxOpenConns(20)
				sqlDB.SetMaxIdleConns(5)
				sqlDB.SetConnMaxLifetime(30 * time.Minute)
				return db, nil
			}
			err = pingErr
		}
		log.Printf("database not ready (attempt %d): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("connect to database: %w", err)
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Store{},
		&models.Product{},
		&models.ShoppingSession{},
		&models.SessionItem{},
		&models.SessionItemOption{},
	)
}
