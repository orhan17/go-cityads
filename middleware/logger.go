package middleware

import (
	"geo_offers/config"
	"geo_offers/models"
	"github.com/gofiber/fiber/v2"
)

// RequestLogger middleware записывает информацию о запросе в БД и в лог-файл.
func RequestLogger(c *fiber.Ctx) error {
	err := c.Next()

	logEntry := models.RequestLog{
		Method:     c.Method(),
		Endpoint:   c.Path(),
		IP:         c.IP(),
		UserAgent:  c.Get("User-Agent"),
		StatusCode: c.Response().StatusCode(),
	}

	if result := config.DB.Create(&logEntry); result.Error != nil {
		config.Logger.Printf("❌ Ошибка сохранения запроса в БД: %v", result.Error)
	}

	config.Logger.Printf("Request: %s %s from %s - %d", c.Method(), c.Path(), c.IP(), c.Response().StatusCode())

	return err
}
