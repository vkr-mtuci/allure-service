package test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vkr-mtuci/allure-service/internal/adapter"
	"github.com/vkr-mtuci/allure-service/internal/handler"
)

// ✅ Тест для `GetNextLaunch`
func TestGetNextLaunchHandler(t *testing.T) {
	mockService := new(MockAllureService)
	app := fiber.New()
	h := handler.NewAllureHandler(mockService)
	app.Get("/next-launch", h.GetNextLaunch)

	// 🛠 Мокируем успешный ответ
	mockLaunch := &adapter.Launch{
		ID:          123,
		Name:        "Test Run",
		CreatedDate: time.Now().UnixMilli(),
	}
	mockService.On("GetNextLaunch", mock.Anything).Return(mockLaunch, nil)

	// 🏃‍♂️ Выполняем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/next-launch?after=2024-02-01T12:00:00Z", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mockService.AssertExpectations(t) // Проверка вызова мока
}

// ✅ Тест для ошибки (когда нет подходящего запуска)
func TestGetNextLaunchHandler_NotFound(t *testing.T) {
	mockService := new(MockAllureService)
	app := fiber.New()
	h := handler.NewAllureHandler(mockService)
	app.Get("/next-launch", h.GetNextLaunch)

	// 🛠 Мокируем ошибку "не найден запуск"
	mockService.On("GetNextLaunch", mock.Anything).Return(nil, errors.New("не найден запуск"))

	// 🏃‍♂️ Выполняем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/next-launch?after=2024-02-01T12:00:00Z", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	mockService.AssertExpectations(t)
}

func TestGeneratePDFReportHandler_InvalidInput(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAllureService)
	handler := handler.NewAllureHandler(mockService)
	app.Post("/export/pdf/:id", handler.GeneratePDFReport)

	// Ожидаем вызов `GeneratePDFReport` с `launchId=123` и `name="Test"`
	mockService.On("GeneratePDFReport", int64(123), "Test").Return(nil, errors.New("invalid input"))

	// Тест с несоответствующим ID в пути и теле
	reqBody := `{"launchId": 123, "name": "Test"}`
	req := httptest.NewRequest("POST", "/export/pdf/456", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	// Проверяем, что обработчик вернул ошибку 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Проверяем, что мок был вызван
	mockService.AssertExpectations(t)
}

func TestDownloadPDFReport_StatusHandling(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAllureService)
	handler := handler.NewAllureHandler(mockService)
	app.Get("/export/pdf/download/:id", handler.DownloadPDFReport)

	// Тест с несуществующим отчетом
	mockService.On("DownloadPDFReport", "999").Return(
		nil,
		"",
		errors.New("report not found"),
	)

	req := httptest.NewRequest("GET", "/export/pdf/download/999", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetNextLaunch_InvalidDateFormat(t *testing.T) {
	app := fiber.New()
	mockService := new(MockAllureService)
	handler := handler.NewAllureHandler(mockService)
	app.Get("/next-launch", handler.GetNextLaunch)

	// Неправильный формат даты
	req := httptest.NewRequest(http.MethodGet, "/next-launch?after=2024-02-01", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetPDFDownloadLink_Handler(t *testing.T) {
	// Создаём мок сервиса
	mockService := new(MockAllureService)
	mockService.On("GetPDFDownloadLink", "456").Return("http://mocked.url/download/456")

	// Создаём обработчик и роутер Fiber
	app := fiber.New()
	handler := handler.NewAllureHandler(mockService)
	app.Get("/export/download/:id", handler.GetPDFDownloadLink)

	// Выполняем HTTP-запрос
	req := httptest.NewRequest("GET", "/export/download/456", nil)
	resp, err := app.Test(req)

	// Проверяем, что запрос выполнен успешно
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем тело ответа
	body, _ := io.ReadAll(resp.Body)
	expected := `{"download_link":"http://mocked.url/download/456"}`
	assert.JSONEq(t, expected, string(body))
}
