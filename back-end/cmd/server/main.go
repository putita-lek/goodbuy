package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/putita-lek/goodbuy-backend/internal/config"
	"github.com/putita-lek/goodbuy-backend/internal/database"
	"github.com/putita-lek/goodbuy-backend/internal/handler"
	"github.com/putita-lek/goodbuy-backend/internal/repository"
	"github.com/putita-lek/goodbuy-backend/internal/seed"
	"github.com/putita-lek/goodbuy-backend/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if cfg.SeedData {
		if err := seed.Run(db); err != nil {
			log.Fatalf("seed: %v", err)
		}
	}

	svc := service.NewShoppingService(repository.NewProductRepository(db), repository.NewSessionRepository(db))

	app := fiber.New(fiber.Config{AppName: "GoodBuy API", BodyLimit: 1 << 20})
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{AllowOrigins: cfg.CORSOrigins}))
	handler.NewShoppingHandler(svc).Register(app)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Println("shutting down")
		_ = app.Shutdown()
	}()

	log.Printf("GoodBuy API listening on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
