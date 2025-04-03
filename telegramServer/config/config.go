package config

import (
	"log"
	"os"
)

type Config struct {
	TelegramToken string
	GRPCHost      string
	GRPCPort      string
}

func LoadConfig() Config {
	return Config{
		TelegramToken: mustGetEnv("TELEGRAM_TOKEN"),
		GRPCHost:      mustGetEnv("GRPC_SERVER_HOST"),
		GRPCPort:      mustGetEnv("GRPC_SERVER_PORT"),
	}
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Переменная окружения %s не установлена", key)
	}
	return val
}
