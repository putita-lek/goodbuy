// Command seed creates the tables if needed and runs the seed, without starting
// the API. It runs regardless of SEED_DATA.
//
// Like the seed at server start, it adds retailers and stores only when the
// retailers table is empty, and mock products only to retailers with none, so
// running it twice adds nothing the second time. Use reset-db first to start over.
package main

import (
	"log"

	"github.com/putita-lek/goodbuy-backend/internal/config"
	"github.com/putita-lek/goodbuy-backend/internal/database"
	"github.com/putita-lek/goodbuy-backend/internal/seed"
)

func main() {
	db, err := database.Connect(config.Load().DSN())
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := seed.Run(db); err != nil {
		log.Fatalf("seed: %v", err)
	}
	log.Printf("seed finished")
}
