package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/vkr-mtuci/allure-service/internal/service"
)

// AllureHandler - обработчик запросов к Allure
type AllureHandler struct {
	service service.AllureServiceInterface
}

// NewAllureHandler - конструктор обработчика
func NewAllureHandler(service service.AllureServiceInterface) *AllureHandler {
	return &AllureHandler{service: service}
}

// GetNextLaunch - обрабатывает запрос поиска следующего запуска после указанной даты
func (h *AllureHandler) GetNextLaunch(c *fiber.Ctx) error {
	dateParam := c.Query("after")
	if dateParam == "" {
		log.Warn().Msg("⚠️ Параметр 'after' не указан")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Необходимо передать параметр 'after' (в формате RFC3339)",
		})
	}

	// 🛠 Заменяем пробел на `+`, если браузер или cURL его заменили
	correctedDate := strings.ReplaceAll(dateParam, " ", "+")

	afterDate, err := time.Parse(time.RFC3339, correctedDate)
	if err != nil {
		log.Error().Err(err).Msgf("❌ Ошибка парсинга даты: %s", correctedDate)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Некорректный формат даты, используйте RFC3339 (например, 2025-01-30T22:00:38.625+03:00)",
		})
	}

	nextLaunch, err := h.service.GetNextLaunch(afterDate)
	if err != nil {
		log.Error().Err(err).Msg("❌ Ошибка при поиске следующего запуска")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(nextLaunch)
}

// GeneratePDFReport - инициирует создание PDF-отчета
func (h *AllureHandler) GeneratePDFReport(c *fiber.Ctx) error {
	var request struct {
		LaunchID        int64  `json:"launchId"`
		Name            string `json:"name"`
		WithPageNumbers bool   `json:"withPageNumbers"`
	}

	// Распарсим JSON из тела запроса
	if err := c.BodyParser(&request); err != nil {
		log.Error().Err(err).Msg("❌ Ошибка парсинга тела запроса")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Некорректный формат JSON",
		})
	}

	// Проверяем, переданы ли все обязательные параметры
	if request.LaunchID == 0 {
		log.Warn().Msg("⚠️ Не указан LaunchID")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Необходимо передать ID запуска",
		})
	}

	if request.Name == "" {
		log.Warn().Msg("⚠️ Не указано имя запуска")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Необходимо передать имя запуска",
		})
	}

	// Вызываем сервис для генерации PDF
	report, err := h.service.GeneratePDFReport(request.LaunchID, request.Name)
	if err != nil {
		log.Error().Err(err).Msg("❌ Ошибка генерации PDF-отчета")

		// Если ошибка связана с валидацией данных, возвращаем 400
		if strings.Contains(err.Error(), "invalid input") {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// В остальных случаях - 500
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка сервера при генерации отчета",
		})
	}

	// ✅ Исправлено: Преобразуем report.ID (int64) в строку перед вызовом GetPDFDownloadLink
	reportIDStr := strconv.FormatInt(report.ID, 10)

	// Отправляем ответ с ID отчета
	return c.JSON(fiber.Map{
		"report_id":     report.ID,
		"download_link": h.service.GetPDFDownloadLink(reportIDStr), // ✅ Теперь строка
	})
}

// GetPDFDownloadLink - получает ссылку на скачивание PDF-отчета
func (h *AllureHandler) GetPDFDownloadLink(c *fiber.Ctx) error {
	reportID := c.Params("id")
	if reportID == "" {
		log.Warn().Msg("⚠️ Не указан reportID")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Необходимо передать ID отчета",
		})
	}

	downloadLink := h.service.GetPDFDownloadLink(reportID)

	return c.JSON(fiber.Map{"download_link": downloadLink})
}

// DownloadPDFReport - скачивает PDF-отчет и передает его на фронт
func (h *AllureHandler) DownloadPDFReport(c *fiber.Ctx) error {
	reportID := c.Params("id")
	if reportID == "" {
		log.Warn().Msg("⚠️ Не указан reportID")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Необходимо передать ID отчета",
		})
	}

	// Запрашиваем скачивание PDF
	fileData, fileName, err := h.service.DownloadPDFReport(reportID)
	if err != nil {
		log.Error().Err(err).Msg("❌ Ошибка скачивания PDF")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка скачивания PDF-отчета",
		})
	}

	// Возвращаем PDF-файл как поток
	c.Set("Content-Disposition", "attachment; filename="+fileName)
	c.Set("Content-Type", "application/pdf")
	return c.Send(fileData)
}
