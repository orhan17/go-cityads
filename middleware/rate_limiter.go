package middleware

import (
	"context"
	"fmt"
	"time"

	"geo_offers/config"
	"github.com/gofiber/fiber/v2"
)

// RateLimiter - ограничение запросов (100 запросов в минуту)
func RateLimiter(c *fiber.Ctx) error {
	ip := c.IP()
	key := fmt.Sprintf("ratelimit:%s", ip)

	count, err := config.RedisClient.Incr(context.Background(), key).Result()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка сервера Redis"})
	}

	if count == 1 {
		config.RedisClient.Expire(context.Background(), key, time.Minute)
	}

	// Если лимит превышен, блокируем, ставим лимит на 30 (можно и побольше)
	if count > 30 {
		return c.Status(429).JSON(fiber.Map{"error": "Слишком много запросов. Попробуйте позже."})
	}

	return c.Next()
}
