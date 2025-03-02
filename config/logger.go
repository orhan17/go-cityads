package config

import (
	"log"
	"os"
)

var Logger *log.Logger

func SetupLogger() {
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Ошибка создания директории для логов: %v", err)
	}

	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Ошибка создания файла логов: %v", err)
	}

	Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	Logger.Println("Логирование запущено")
}
