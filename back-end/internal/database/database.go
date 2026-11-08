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
	if err := dropLegacyCatalogue(db); err != nil {
		return err
	}
	return db.AutoMigrate(
		&models.Retailer{},
		&models.Store{},
		&models.Product{},
		&models.ProductPrice{},
		&models.ShoppingSession{},
		&models.SessionItem{},
		&models.SessionItemOption{},
	)
}

// dropLegacyCatalogue removes catalogue tables from earlier designs, which
// AutoMigrate cannot convert (it never drops old columns or tables):
//   - stores holding the brand name, with each product belonging to one store;
//   - store_products, which held a price per store before prices moved to product_prices.
//
// Retailers and stores of the current design are kept. Products are dropped and
// must be imported again. Session tables are kept: they hold snapshots, not references.
func dropLegacyCatalogue(db *gorm.DB) error {
	m := db.Migrator()
	if (m.HasTable("stores") && m.HasColumn("stores", "name")) ||
		(m.HasTable("products") && m.HasColumn("products", "store_id")) {
		log.Printf("dropping old stores/products tables to migrate to the retailer/store design")
		return m.DropTable("store_products", "product_prices", "products", "stores")
	}
	if m.HasTable("store_products") {
		log.Printf("dropping store_products and products: prices now live in product_prices, re-run the import")
		return m.DropTable("store_products", "product_prices", "products")
	}
	return nil
}
