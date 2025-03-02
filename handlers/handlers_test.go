// handlers_test.go
package handlers_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9" // Используем Redis v9
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"geo_offers/config"
	"geo_offers/handlers"
	"geo_offers/models"
)

// setupTestEnv подготавливает тестовую среду: in-memory SQLite, miniredis и Fiber-приложение с маршрутами, как в main.go.
func setupTestEnv(t *testing.T) *fiber.App {
	// Настраиваем in-memory SQLite базу
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	config.DB = db

	// Применяем миграцию для модели Offer
	err = config.DB.AutoMigrate(&models.Offer{})
	assert.NoError(t, err)

	// Настраиваем fake Redis через miniredis
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	config.RedisClient = rdb

	// Создаем Fiber-приложение и регистрируем маршруты согласно main.go
	app := fiber.New()
	app.Get("/api/v1/ping", handlers.Ping)
	app.Get("/api/v1/health", handlers.HealthCheck)
	app.Get("/api/v1/offers/:geo", handlers.GetOffersByGeo)
	app.Get("/api/v1/geo-stats", handlers.GetGeoStats)
	app.Get("/api/v1/offers-sorted", handlers.GetAllOffersSortedByRating)
	app.Post("/offers", handlers.CreateOffer)

	return app
}

// TestPing проверяет работу обработчика Ping.
func TestPing(t *testing.T) {
	app := setupTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/ping", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "pong", string(body))
}

// TestHealthCheck проверяет обработчик HealthCheck.
func TestHealthCheck(t *testing.T) {
	app := setupTestEnv(t)

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var response map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

// TestGetOffersByGeo проверяет обработчик получения офферов по GEO.
func TestGetOffersByGeo(t *testing.T) {
	app := setupTestEnv(t)

	// Добавляем тестовый оффер. ExternalID – числовое значение.
	offer := models.Offer{
		GeoCode:    "RU",
		ExternalID: 123,
		Rating:     5,
	}
	result := config.DB.Create(&offer)
	assert.NoError(t, result.Error)

	req := httptest.NewRequest("GET", "/api/v1/offers/RU?page=1&limit=5", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	offers, ok := response["offers"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, offers)
}

// TestGetGeoStats проверяет обработчик получения статистики по GEO.
func TestGetGeoStats(t *testing.T) {
	app := setupTestEnv(t)

	// Добавляем тестовые офферы для разных GEO.
	offers := []models.Offer{
		{GeoCode: "RU", ExternalID: 1, Rating: 5},
		{GeoCode: "RU", ExternalID: 2, Rating: 4},
		{GeoCode: "US", ExternalID: 3, Rating: 3},
	}
	for _, o := range offers {
		result := config.DB.Create(&o)
		assert.NoError(t, result.Error)
	}

	req := httptest.NewRequest("GET", "/api/v1/geo-stats", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var stats []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&stats)
	assert.NoError(t, err)
	// Ожидаем минимум 2 записи (например, для RU и US)
	assert.GreaterOrEqual(t, len(stats), 2)
}

// TestGetAllOffersSortedByRating проверяет получение офферов, отсортированных по рейтингу.
func TestGetAllOffersSortedByRating(t *testing.T) {
	app := setupTestEnv(t)

	// Добавляем тестовые офферы с разными рейтингами.
	offers := []models.Offer{
		{GeoCode: "RU", ExternalID: 1, Rating: 2},
		{GeoCode: "RU", ExternalID: 2, Rating: 5},
		{GeoCode: "RU", ExternalID: 3, Rating: 3},
	}
	for _, o := range offers {
		result := config.DB.Create(&o)
		assert.NoError(t, result.Error)
	}

	req := httptest.NewRequest("GET", "/api/v1/offers-sorted?page=1&limit=5", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	offersResp, ok := response["offers"].([]interface{})
	assert.True(t, ok)
	if len(offersResp) >= 2 {
		first := offersResp[0].(map[string]interface{})
		second := offersResp[1].(map[string]interface{})
		firstRating := int(first["rating"].(float64))
		secondRating := int(second["rating"].(float64))
		assert.True(t, firstRating >= secondRating)
	}
}

// TestCreateOffer проверяет создание нового оффера.
func TestCreateOffer(t *testing.T) {
	app := setupTestEnv(t)

	// Устанавливаем тестовый API-токен.
	os.Setenv("API_TOKEN", "test-token")

	// JSON для создания оффера. ExternalID теперь числовое.
	offerData := `{"external_id": 100, "geo_code": "RU", "rating": 4}`
	req := httptest.NewRequest("POST", "/offers", strings.NewReader(offerData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "test-token")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 201, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Оффер создан успешно", response["message"])
}
