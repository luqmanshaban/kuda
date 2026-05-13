package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost string 
	DBPort int 
	DBUser string
	DBName string 
	DBPassword string 
	Port string 
	APIKey string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnvAsInt("DB_PORT", 5432),
		DBUser: getEnv("DB_USER", "postgres"),
		DBName: getEnv("DB_NAME", "kuda"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		Port: getEnv("PORT", ":8000"),
		APIKey: getEnv("KUDA_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return  defaultValue
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value 
	}
	return defaultValue
}