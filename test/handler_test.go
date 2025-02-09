package test

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
