package main

import (
	"fmt"
	"log"
	"os"

	"geo_offers/config"
	"geo_offers/handlers"
	"geo_offers/middleware"
	"geo_offers/models"
	"geo_offers/services"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// initEnv загружает переменные окружения из .env файла
func initEnv() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Не удалось загрузить .env файл. Используются переменные окружения по умолчанию.")
	}
}

// initConnections настраивает логгер, подключается к базе данных, Redis и выполняет миграции
func initConnections() {
	config.SetupLogger()
	config.ConnectDB()
	config.ConnectRedis()

	if err := config.DB.AutoMigrate(&models.Offer{}); err != nil {
		log.Fatalf("Ошибка миграции: %v", err)
	}
}

// setupRoutes регистрирует все маршруты API
func setupRoutes(app *fiber.App) {
	// Middleware
	app.Use(middleware.RateLimiter)
	app.Use(middleware.RequestLogger)

	// Роуты API
	app.Get("/api/v1/offers/:geo", handlers.GetOffersByGeo)
	app.Get("/api/v1/geo-stats", handlers.GetGeoStats)
	app.Get("/api/v1/offers-sorted", handlers.GetAllOffersSortedByRating)
	app.Get("/api/v1/metrics", middleware.MetricsHandler())
	app.Get("/api/v1/health", handlers.HealthCheck)
	app.Get("/api/v1/ping", handlers.Ping)

	// Роут для запуска синхронизации офферов
	app.Post("/sync-offers", func(c *fiber.Ctx) error {
		go services.SyncOffers()
		return c.JSON(fiber.Map{"message": "Синхронизация запущена"})
	})

	// Роут для создания оффера
	app.Post("/offers", handlers.CreateOffer)
}

// @title Geo Offers API
// @description API для предоставления офферов по GEO. Предоставляет методы для синхронизации, создания и получения офферов.
// @contact.name API Support
// @host localhost:3000
// @BasePath /api/v1
func main() {
	initEnv()
	initConnections()

	// Запуск фоновой синхронизации офферов
	go services.SyncOffers()

	app := fiber.New()
	setupRoutes(app)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Printf("API запущено на порту %s\n", port)
	log.Fatal(app.Listen(":" + port))
}
