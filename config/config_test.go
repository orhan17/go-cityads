package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"geo_offers/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

// TestSetupLogger проверяет, что SetupLogger создаёт директорию logs и файл logs/app.log, а также инициализирует config.Logger.
func TestSetupLogger(t *testing.T) {
	// Удаляем директорию logs, если она существует, чтобы проверить создание заново.
	os.RemoveAll("logs")

	config.SetupLogger()

	// Проверяем, что config.Logger не nil.
	assert.NotNil(t, config.Logger)

	// Проверяем, что директория logs существует.
	info, err := os.Stat("logs")
	assert.NoError(t, err)
	assert.True(t, info.IsDir())

	// Проверяем, что файл logs/app.log существует.
	logFilePath := filepath.Join("logs", "app.log")
	info, err = os.Stat(logFilePath)
	assert.NoError(t, err)
	assert.False(t, info.IsDir())
}

// TestConnectRedis проверяет, что ConnectRedis устанавливает соединение с Redis.
// Для эмуляции Redis используется miniredis.
func TestConnectRedis(t *testing.T) {
	// Запускаем miniredis.
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	// Устанавливаем переменную окружения REDIS_HOST на адрес miniredis.
	os.Setenv("REDIS_HOST", mr.Addr())

	// Вызываем функцию подключения.
	config.ConnectRedis()

	// Проверяем, что config.RedisClient не nil.
	assert.NotNil(t, config.RedisClient)

	// Проверяем, что Ping возвращает "PONG".
	ctx := context.Background()
	pong, err := config.RedisClient.Ping(ctx).Result()
	assert.NoError(t, err)
	assert.Equal(t, "PONG", pong)
}

// TestConnectDB пытается установить соединение с MySQL.
func TestConnectDB(t *testing.T) {
	t.Skip("Skipping TestConnectDB")
}
