package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config - структура для хранения конфигурации приложения
type Config struct {
	ServerPort      string
	AllureBaseURL   string
	AllureAPIURL    string
	AllureUserToken string
	AllureProjectID string
	TokenExpiry     time.Duration
}

// LoadConfig загружает переменные окружения в структуру Config
func LoadConfig() *Config {
	envPath := ".env"
	if err := godotenv.Load(envPath); err != nil {
		log.Println("⚠ Нет .env файла, используем переменные окружения")
	}

	config := &Config{
		ServerPort:      os.Getenv("SERVER_PORT"),
		AllureBaseURL:   os.Getenv("ALLURE_BASE_URL"),
		AllureAPIURL:    os.Getenv("ALLURE_API_URL"),
		AllureUserToken: os.Getenv("ALLURE_API_TOKEN"),
		AllureProjectID: os.Getenv("ALLURE_PROJECT_ID"),
		TokenExpiry:     55 * time.Minute, // Настройка истечения токена
	}

	if config.AllureBaseURL == "" || config.AllureUserToken == "" || config.AllureAPIURL == "" || config.AllureProjectID == "" {
		log.Fatal("❌ Ошибка: Не заданы все обязательные переменные окружения для Allure")
	}

	return config
}
