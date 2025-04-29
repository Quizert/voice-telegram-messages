package config

import (
	"log"
	"os"
)

type Config struct {
	TelegramToken string
	GRPCHost      string
	GRPCPort      string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
}

func LoadConfig() Config {
	return Config{
		TelegramToken: mustGetEnv("TELEGRAM_TOKEN"),
		GRPCHost:      mustGetEnv("GRPC_SERVER_HOST"),
		GRPCPort:      mustGetEnv("GRPC_SERVER_PORT"),
		DBHost:        mustGetEnv("DB_HOST"),
		DBPort:        mustGetEnv("DB_PORT"),
		DBUser:        mustGetEnv("DB_USER"),
		DBPassword:    mustGetEnv("DB_PASSWORD"),
		DBName:        mustGetEnv("DB_NAME"),
	}
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Переменная окружения %s не установлена", key)
	}
	return val
}
