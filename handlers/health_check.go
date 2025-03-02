package handlers

import (
	"geo_offers/config"
	"github.com/gofiber/fiber/v2"
)

// HealthCheck проверяет доступность основных зависимостей, например, базы данных
func HealthCheck(c *fiber.Ctx) error {
	// Проверка подключения к базе данных
	sqlDB, err := config.DB.DB()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Ошибка получения подключения к БД",
		})
	}
	if err = sqlDB.Ping(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Пинг БД не прошёл",
		})
	}

	// При желании можно добавить проверку других сервисов (например, Redis)
	// redisErr := config.Redis.Ping().Err() ...

	return c.JSON(fiber.Map{"status": "ok"})
}

// Ping просто отвечает "pong" для проверки доступности сервера
func Ping(c *fiber.Ctx) error {
	return c.SendString("pong")
}
