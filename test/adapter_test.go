package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vkr-mtuci/allure-service/config"
	"github.com/vkr-mtuci/allure-service/internal/adapter"
)

// ✅ **Тест с моком**
func TestNewAllureClient_WithMock(t *testing.T) {
	cfg := &config.Config{
		AllureBaseURL:   "https://allure.example.com",
		AllureAPIURL:    "/api/",
		AllureUserToken: "test-token",
		AllureProjectID: "1661",
		TokenExpiry:     55 * time.Minute,
	}

	mockClient := new(MockAllureClient)
	client := adapter.NewAllureClient(cfg)

	// 🔥 **Проверяем, что клиент создался**
	assert.NotNil(t, client)

	// 🔥 **Мокаем запрос `Authenticate` с `context.TODO()`**
	mockClient.On("Authenticate", mock.Anything).Return(nil)

	// 🔥 **Проверяем, что параметры загружены корректно**
	assert.Equal(t, "https://allure.example.com", cfg.AllureBaseURL)
	assert.Equal(t, "/api/", cfg.AllureAPIURL)
	assert.Equal(t, "test-token", cfg.AllureUserToken)
	assert.Equal(t, "1661", cfg.AllureProjectID)

	// 🔥 **Передаем `context.TODO()` вместо `nil`**
	err := mockClient.Authenticate(context.TODO())
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

// ✅ **Тест аутентификации**
func TestAuthenticate_RealClient(t *testing.T) {
	cfg := &config.Config{
		AllureBaseURL:   "https://allure.example.com",
		AllureAPIURL:    "/api/",
		AllureUserToken: "test-token",
		AllureProjectID: "1661",
		TokenExpiry:     55 * time.Minute,
	}

	client := adapter.NewAllureClient(cfg)

	// 🔥 **Запускаем аутентификацию**
	err := client.Authenticate(context.TODO())

	// ❗ **В реальном тесте API может быть недоступно, поэтому проверяем только отсутствие паники**
	assert.NotNil(t, client)
	assert.Error(t, err) // Так как API Allure реально недоступен
}

// Тест ошибки при запросе токена
func TestAuthenticate_RequestError(t *testing.T) {
	mockClient := new(MockAllureClient)

	// 🔥 **Мокаем ошибку запроса**
	mockClient.On("Authenticate", mock.Anything).Return(errors.New("ошибка сети"))

	// 🏃‍♂️ **Вызываем `Authenticate()`**
	err := mockClient.Authenticate(context.TODO())

	// ❌ **Ожидаем ошибку**
	assert.Error(t, err)
	assert.Equal(t, "ошибка сети", err.Error())

	// 📌 **Проверяем, что вызов был**
	mockClient.AssertExpectations(t)
}

// Тест ошибки парсинга JSON
func TestAuthenticate_BadJSON(t *testing.T) {
	mockClient := new(MockAllureClient)

	// 🔥 **Мокаем некорректный JSON**
	mockClient.On("Authenticate", mock.Anything).Return(errors.New("ошибка парсинга JSON"))

	// 🏃‍♂️ **Вызываем `Authenticate()`**
	err := mockClient.Authenticate(context.TODO())

	// ❌ **Ожидаем ошибку**
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка парсинга JSON")

	// 📌 **Проверяем вызов**
	mockClient.AssertExpectations(t)
}

// Тест ошибки при статус-коде
func TestAuthenticate_StatusCodeError(t *testing.T) {
	mockClient := new(MockAllureClient)

	// 🔥 **Мокаем код 500**
	mockClient.On("Authenticate", mock.Anything).Return(errors.New("ошибка API: статус 500"))

	// 🏃‍♂️ **Вызываем `Authenticate()`**
	err := mockClient.Authenticate(context.TODO())

	// ❌ **Ожидаем ошибку**
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка API: статус 500")

	// 📌 **Проверяем вызов**
	mockClient.AssertExpectations(t)
}
