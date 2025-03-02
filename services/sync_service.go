package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"context"
	"geo_offers/config"
	"geo_offers/models"
	"github.com/go-resty/resty/v2"
)

const maxPages = 100

// SyncOffers @Summary Synchronize offers from external API
// @Description Loads offers from an external API , updates or creates offers in the database, and clears cache for updated GEO codes.
// @Tags Sync
// @Produce plain
// @Success 200 {string} string "Все офферы загружены, обновлены и кеш очищен!"
func SyncOffers() {
	client := resty.New()
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	baseURL := os.Getenv("API_URL")

	updatedOffers := make(map[int]bool)
	newOffers := make(map[int]bool)

	// Это нам нужен чтобы кеш удалять (удалять которые обновились)
	geoUpdated := make(map[string]bool)

	// Данные загружаем
	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s?page=%d", baseURL, page)
		resp, err := client.R().Get(url)
		if err != nil {
			log.Println("Ошибка запроса к API:", err)
			break
		}

		var apiResponse APIResponse
		if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
			log.Println("Ошибка парсинга JSON:", err)
			break
		}

		if len(apiResponse.Offers) == 0 {
			fmt.Println("Достигнут конец страниц. Синхронизация завершена.")
			break
		}

		for _, extOffer := range apiResponse.Offers {
			if len(extOffer.Geo) == 0 {
				continue
			}

			externalID, err := strconv.Atoi(extOffer.ExternalID)
			if err != nil {
				log.Printf("Ошибка конвертации ExternalID: %s\n", extOffer.ExternalID)
				continue
			}

			approvalTime, _ := strconv.Atoi(extOffer.ApprovalTime)
			paymentTime, _ := strconv.Atoi(extOffer.PaymentTime)
			ecpl, _ := strconv.ParseFloat(extOffer.Stat.ECPL, 64)

			for _, geo := range extOffer.Geo {
				if geo.Code == "Wrld" {
					continue
				}

				// Здесь вычисляем рейтинг
				rating := ecpl * (10 * (1 - float64(approvalTime)/90)) * (100 * (1 - float64(paymentTime)/90))

				newOffer := models.Offer{
					ExternalID:   externalID,
					Name:         extOffer.Name,
					Currency:     extOffer.OfferCurrency.Name,
					ApprovalTime: approvalTime,
					SiteURL:      extOffer.SiteURL,
					Logo:         extOffer.Logo,
					GeoCode:      geo.Code,
					GeoName:      geo.Name,
					Rating:       rating,
				}

				// Здеьс проверяем есть ли оффер в БД
				var existingOffer models.Offer
				result := config.DB.Where("external_id = ?", externalID).First(&existingOffer)

				if result.RowsAffected > 0 {
					// Если оффер найден -> обновляем данные
					config.DB.Model(&existingOffer).Updates(newOffer)
					updatedOffers[externalID] = true
					geoUpdated[geo.Code] = true // Ставим true чтобы обновить кеш для этого гео кода
				} else {
					// Если оффера нет -> создаем новый
					config.DB.Create(&newOffer)
					newOffers[externalID] = true
					geoUpdated[geo.Code] = true // тоже самое тут делаем
				}
			}
		}
	}

	// Здесь очищаем кеш
	for geo := range geoUpdated {
		clearCacheByGeo(geo)
	}

	if len(updatedOffers) > 0 {
		fmt.Printf("Обновлены %d офферов: %v\n", len(updatedOffers), getKeys(updatedOffers))
	}
	if len(newOffers) > 0 {
		fmt.Printf("Добавлены %d новых офферов: %v\n", len(newOffers), getKeys(newOffers))
	}

	fmt.Println("Все офферы загружены, обновлены и кеш очищен!")
}

func getKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Вспомогательная функция для того чтобы очистить кеш
func clearCacheByGeo(geo string) {
	cacheKeyPattern := fmt.Sprintf("offers:%s:*", geo)
	config.RedisClient.Del(context.Background(), cacheKeyPattern)
	fmt.Println("Кеш очищен для GEO:", geo)
}
