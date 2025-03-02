package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"geo_offers/config"
	"geo_offers/models"
	"github.com/gofiber/fiber/v2"
	"os"
	"strconv"
	"time"
)

// GetOffersByGeo godoc
// @Summary Получение офферов по GEO
// @Description Возвращает офферы для указанного GEO с пагинацией и кешированием.
// @Tags Offers
// @Accept json
// @Produce json
// @Param geo path string true "GEO код"
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество записей на страницу" default(5)
// @Success 200 {object} fiber.Map
// @Failure 404 {object} fiber.Map{"error": "Офферы для данного ГЕО не найдены"}
// @Router /offers/{geo} [get]
func GetOffersByGeo(c *fiber.Ctx) error {
	geo := c.Params("geo")

	// Здесь получаем параметры пагинации
	limit, err := strconv.Atoi(c.Query("limit", "5"))
	if err != nil || limit < 1 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	// Здесь генерируем ключ для кеша
	cacheKey := fmt.Sprintf("offers:%s:page:%d:limit:%d", geo, page, limit)

	// Проверка кеша
	cachedData, err := config.RedisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		fmt.Println("Данные загружены из кеша")
		return c.SendString(cachedData)
	}

	// Здесь данные качаем из БД
	var offers []models.Offer
	config.DB.Where("geo_code = ?", geo).Order("rating DESC").Limit(limit).Offset(offset).Find(&offers)

	var total int64
	config.DB.Model(&models.Offer{}).Where("geo_code = ?", geo).Count(&total)

	if len(offers) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Офферы для данного ГЕО не найдены"})
	}

	response := fiber.Map{
		"total":       total,
		"limit":       limit,
		"page":        page,
		"total_pages": (int(total) + limit - 1) / limit,
		"offers":      offers,
	}

	// Здесь сохраняем кеш на 10 минут
	data, _ := json.Marshal(response)
	config.RedisClient.Set(context.Background(), cacheKey, data, 10*time.Minute)

	return c.JSON(response)
}

// GetGeoStats godoc
// @Summary Получение статистики по GEO
// @Description Возвращает статистику офферов для каждого GEO.
// @Tags Offers
// @Produce json
// @Success 200 {object} []struct{GeoCode string; Count int}
// @Router /geo-stats [get]
func GetGeoStats(c *fiber.Ctx) error {
	var stats []struct {
		GeoCode string `json:"geo_code"`
		Count   int    `json:"count"`
	}

	config.DB.Raw("SELECT geo_code, COUNT(*) as count FROM offers GROUP BY geo_code").Scan(&stats)

	return c.JSON(stats)
}

// GetAllOffersSortedByRating godoc
// @Summary Получение всех офферов, отсортированных по рейтингу
// @Description Возвращает все офферы, отсортированные по убыванию рейтинга, с пагинацией и кешированием.
// @Tags Offers
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Количество записей на страницу" default(5)
// @Success 200 {object} fiber.Map
// @Failure 404 {object} fiber.Map{"error": "Офферы не найдены"}
// @Router /offers-sorted [get]
func GetAllOffersSortedByRating(c *fiber.Ctx) error {
	// Получаем параметры пагинации
	limit, err := strconv.Atoi(c.Query("limit", "5"))
	if err != nil || limit < 1 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	// Здесь генерируем ключ для кеша
	cacheKey := fmt.Sprintf("offers_sorted:page:%d:limit:%d", page, limit)

	// Проверка кеша
	cachedData, err := config.RedisClient.Get(context.Background(), cacheKey).Result()
	if err == nil {
		fmt.Println("Данные загружены из кеша")
		return c.SendString(cachedData)
	}

	var offers []models.Offer

	config.DB.Order("rating DESC").Limit(limit).Offset(offset).Find(&offers)

	var total int64
	config.DB.Model(&models.Offer{}).Count(&total)

	if len(offers) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Офферы не найдены"})
	}

	response := fiber.Map{
		"total":       total,
		"limit":       limit,
		"page":        page,
		"total_pages": (int(total) + limit - 1) / limit,
		"offers":      offers,
	}

	// Здесь сохраняем кеш на 10 минут
	data, _ := json.Marshal(response)
	config.RedisClient.Set(context.Background(), cacheKey, data, 10*time.Minute)

	return c.JSON(response)
}

// CreateOffer godoc
// @Summary Создание нового оффера
// @Description Создает новый оффер. Требует авторизации через API-токен.
// @Tags Offers
// @Accept json
// @Produce json
// @Param offer body models.Offer true "Параметры оффера"
// @Success 201 {object} fiber.Map
// @Failure 400 {object} fiber.Map{"error": "Ошибка парсинга данных"}
// @Failure 401 {object} fiber.Map{"error": "Доступ запрещён. Неверный API-токен."}
// @Failure 409 {object} fiber.Map{"error": "Оффер с таким ExternalID уже существует"}
// @Router /offers [post]
func CreateOffer(c *fiber.Ctx) error {
	// Здесь проверяем апи токен (самая простая реализация)
	apiToken := c.Get("Authorization")
	expectedToken := os.Getenv("API_TOKEN")

	if apiToken == "" || apiToken != expectedToken {
		return c.Status(401).JSON(fiber.Map{"error": "Доступ запрещён. Неверный API-токен."})
	}

	var offer models.Offer

	if err := c.BodyParser(&offer); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Ошибка парсинга данных"})
	}

	var existingOffer models.Offer
	result := config.DB.Where("external_id = ?", offer.ExternalID).First(&existingOffer)

	if result.RowsAffected > 0 {
		return c.Status(409).JSON(fiber.Map{"error": "Оффер с таким ExternalID уже существует"})
	}

	// Здесь сохраняем оффер
	config.DB.Create(&offer)

	return c.Status(201).JSON(fiber.Map{
		"message": "Оффер создан успешно",
		"offer":   offer,
	})
}
