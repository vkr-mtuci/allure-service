package test

import (
	"context"
	"errors"
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
	"github.com/vkr-mtuci/allure-service/internal/service"
)

// ✅ **Тест интеграции `GetNextLaunch`**
func TestIntegrationGetNextLaunch(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)
	handler := handler.NewAllureHandler(service)
	app := fiber.New()
	app.Get("/next-launch", handler.GetNextLaunch)

	// 📌 **Добавляем мок `GetLaunches()`**
	mockLaunches := []adapter.Launch{
		{
			ID:          123,
			Name:        "Test Run",
			CreatedDate: time.Now().UnixMilli(),
		},
	}
	mockClient.On("GetLaunches", mock.Anything).Return(mockLaunches, nil)

	// 🏃‍♂️ Выполняем тестовый запрос
	req := httptest.NewRequest(http.MethodGet, "/next-launch?after=2024-02-01T12:00:00Z", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mockClient.AssertExpectations(t) // Проверяем, что мок вызван правильно
}

// ✅ **Тест ошибки, если запусков нет**
func TestIntegrationGetNextLaunch_NotFound(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)
	handler := handler.NewAllureHandler(service)
	app := fiber.New()
	app.Get("/next-launch", handler.GetNextLaunch)

	// 📌 **Мокаем `GetLaunches()` без данных**
	mockClient.On("GetLaunches", mock.Anything).Return([]adapter.Launch{}, nil)

	// 🏃‍♂️ Выполняем тест
	req := httptest.NewRequest(http.MethodGet, "/next-launch?after=2024-02-01T12:00:00Z", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // Ожидаем ошибку 500

	mockClient.AssertExpectations(t)
}

// ✅ **Тест ошибки `GetLaunches()`**
func TestIntegrationGetNextLaunch_Error(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)
	handler := handler.NewAllureHandler(service)
	app := fiber.New()
	app.Get("/next-launch", handler.GetNextLaunch)

	// 📌 **Мокаем ошибку в `GetLaunches()` (возвращаем пустой массив вместо nil!)**
	mockClient.On("GetLaunches", mock.Anything).Return([]adapter.Launch{}, errors.New("ошибка API"))

	// 🏃‍♂️ Выполняем тест
	req := httptest.NewRequest(http.MethodGet, "/next-launch?after=2024-02-01T12:00:00Z", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // Ожидаем 500

	mockClient.AssertExpectations(t)
}

func TestMockGeneratePDFReport(t *testing.T) {
	mockClient := new(MockAllureClient)

	// 🔹 Мокаем `GeneratePDFReport`
	mockClient.On("GeneratePDFReport", mock.Anything, int64(123), "Test Run").
		Return(&adapter.PDFReport{
			ID:          456,
			Name:        "Test Report",
			ProjectID:   1661,
			Status:      "READY",
			CreatedDate: time.Now().UnixMilli(),
		}, nil)

	// 🏃‍♂️ Вызываем `GeneratePDFReport`
	report, err := mockClient.GeneratePDFReport(context.TODO(), 123, "Test Run")

	// ✅ Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, int64(456), report.ID)
	assert.Equal(t, "Test Report", report.Name)

	// ✅ Проверяем, что мок был вызван
	mockClient.AssertExpectations(t)
}

func TestMockGetPDFDownloadLink(t *testing.T) {
	mockClient := new(MockAllureClient)

	// 🔹 Мокаем `GetPDFDownloadLink`
	mockClient.On("GetPDFDownloadLink", "456").Return("https://allure.example.com/download/456")

	// 🏃‍♂️ Вызываем `GetPDFDownloadLink`
	link := mockClient.GetPDFDownloadLink("456")

	// ✅ Проверяем результат
	assert.Equal(t, "https://allure.example.com/download/456", link)

	// ✅ Проверяем, что мок был вызван
	mockClient.AssertExpectations(t)
}

func TestMockDownloadPDFReport(t *testing.T) {
	mockClient := new(MockAllureClient)

	// 🔹 Фейковый PDF-файл
	pdfData := []byte("%PDF-1.4 Mock PDF File")
	fileName := "mock-report.pdf"

	// 🔹 Мокаем `DownloadPDFReport`
	mockClient.On("DownloadPDFReport", mock.Anything, "456").
		Return(pdfData, fileName, nil)

	// 🏃‍♂️ Вызываем `DownloadPDFReport`
	data, name, err := mockClient.DownloadPDFReport(context.TODO(), "456")

	// ✅ Проверяем результат
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, fileName, name)

	// ✅ Проверяем, что мок был вызван
	mockClient.AssertExpectations(t)
}

func TestFullPDFFlow(t *testing.T) {
	// Инициализация приложения
	app := fiber.New()
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)
	handler := handler.NewAllureHandler(service)

	app.Post("/export/pdf/:id", handler.GeneratePDFReport)
	app.Get("/export/pdf/download/:id", handler.DownloadPDFReport)

	// Мокаем успешный поток
	mockClient.On("GeneratePDFReport", mock.Anything, int64(123), "Test Run").
		Return(&adapter.PDFReport{ID: 456}, nil)

	mockClient.On("GetPDFDownloadLink", "456").
		Return("http://mocked.url/download/456")

	mockClient.On("DownloadPDFReport", mock.Anything, "456").
		Return([]byte("PDF content"), "report.pdf", nil)

	// Шаг 1: Генерация отчета
	reqGen := httptest.NewRequest("POST", "/export/pdf/123", strings.NewReader(
		`{"launchId":123,"name":"Test Run"}`,
	))
	reqGen.Header.Set("Content-Type", "application/json")
	respGen, _ := app.Test(reqGen)
	assert.Equal(t, http.StatusOK, respGen.StatusCode)

	// Шаг 2: Скачивание отчета
	reqDown := httptest.NewRequest("GET", "/export/pdf/download/456", nil)
	respDown, _ := app.Test(reqDown)
	assert.Equal(t, http.StatusOK, respDown.StatusCode)
	assert.Equal(t, "application/pdf", respDown.Header.Get("Content-Type"))
}
