package config

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
)

var RedisClient *redis.Client
var ctx = context.Background()

// Здесь подключение Редиса
func ConnectRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Ошибка подключения к Redis:", err)
	} else {
		fmt.Println("Подключение к Redis успешно!")
	}
}
