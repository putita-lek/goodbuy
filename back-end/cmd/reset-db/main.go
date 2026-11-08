// Command reset-db removes all data from every GoodBuy table, so the seed runs
// again on the next server start (it only seeds an empty retailers table).
//
// Tables are emptied with TRUNCATE, which also resets auto-increment ids. The
// tables themselves are kept. Kaggle products must be imported again afterwards.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/putita-lek/goodbuy-backend/internal/config"
	"github.com/putita-lek/goodbuy-backend/internal/database"
	"github.com/putita-lek/goodbuy-backend/internal/models"
	"gorm.io/gorm"
)

// tables lists every model; all of their tables are emptied.
var tables = []any{
	&models.SessionItemOption{},
	&models.SessionItem{},
	&models.ShoppingSession{},
	&models.ProductPrice{},
	&models.Product{},
	&models.Store{},
	&models.Retailer{},
}

func main() {
	yes := flag.Bool("yes", false, "skip the confirmation prompt")
	flag.Parse()

	cfg := config.Load()
	if !*yes {
		fmt.Printf("This deletes ALL data in database %q on %s:%s. Type \"yes\" to continue: ",
			cfg.DBName, cfg.DBHost, cfg.DBPort)
		answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if strings.TrimSpace(answer) != "yes" {
			fmt.Println("Cancelled, nothing was deleted.")
			return
		}
	}

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}

	// Foreign key checks are a per-connection setting, so run everything on one connection.
	err = db.Connection(func(conn *gorm.DB) error {
		if err := conn.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
			return err
		}
		defer conn.Exec("SET FOREIGN_KEY_CHECKS = 1")

		m := conn.Migrator()
		for _, model := range tables {
			stmt := &gorm.Statement{DB: conn}
			if err := stmt.Parse(model); err != nil {
				return err
			}
			name := stmt.Schema.Table
			if !m.HasTable(name) {
				log.Printf("%s: does not exist, skipped", name)
				continue
			}
			if err := conn.Exec("TRUNCATE TABLE " + conn.Statement.Quote(name)).Error; err != nil {
				return fmt.Errorf("truncate %s: %w", name, err)
			}
			log.Printf("%s: emptied", name)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("done: all tables are empty. Restart the server to seed, then re-run the Kaggle import.")
}
