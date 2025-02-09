package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"

	"github.com/vkr-mtuci/allure-service/config"
	"github.com/vkr-mtuci/allure-service/internal/adapter"
	"github.com/vkr-mtuci/allure-service/internal/handler"
	"github.com/vkr-mtuci/allure-service/internal/service"
)

func main() {
	// Настройка логирования
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	logger := zerolog.New(output).With().Timestamp().Logger()

	// Загрузка конфигурации
	cfg := config.LoadConfig()
	logger.Info().Msg("📢 Запуск Allure-сервиса...")

	// Создание клиента
	allureClient := adapter.NewAllureClient(cfg)

	// Создание сервиса
	allureService := service.NewAllureService(allureClient)

	// Создание обработчика
	allureHandler := handler.NewAllureHandler(allureService)

	// Инициализация Fiber
	app := fiber.New()

	// Включение CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Маршруты API
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "✅ Allure-service is running"})
	})

	app.Get("/next-launch", allureHandler.GetNextLaunch)
	app.Post("/export/pdf/:id", allureHandler.GeneratePDFReport)
	app.Get("/export/download/:id", allureHandler.GetPDFDownloadLink)
	app.Get("/export/pdf/download/:id", allureHandler.DownloadPDFReport)

	// Запуск сервера
	logger.Info().Msgf("🚀 Сервис запущен на порту %s", cfg.ServerPort)
	err := app.Listen(":" + cfg.ServerPort)
	if err != nil {
		logger.Fatal().Err(err).Msg("❌ Ошибка запуска сервера")
	}
}
