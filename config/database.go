package config

import (
	"fmt"
	"log"
	"os"

	"geo_offers/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	// Здесь мы миграцию запускаем через Горм
	err = db.AutoMigrate(&models.Offer{})
	err = db.AutoMigrate(&models.RequestLog{})
	if err != nil {
		log.Fatal("Ошибка миграции БД:", err)
	}

	DB = db
	fmt.Println("Подключение к БД установлено и миграция выполнена")
}
