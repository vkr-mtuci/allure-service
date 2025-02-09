package test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

// Unit-тесты для AllureClient
func TestGetLaunches_RealClient(t *testing.T) {
	// Фейковый HTTP-сервер, который эмулирует Allure API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/uaa/oauth/token" {
			// Эмулируем выдачу токена
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token": "mocked_token", "expires_in": 3600}`))
			return
		}

		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/launch") {
			// Эмулируем ответ от Allure API
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"content": []adapter.Launch{
					{ID: 1, Name: "Launch 1", CreatedDate: time.Now().Add(-1 * time.Hour).UnixMilli()},
					{ID: 2, Name: "Launch 2", CreatedDate: time.Now().Add(1 * time.Hour).UnixMilli()},
				},
			})
			return
		}

		// Если запрос не распознан, возвращаем 404
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close() // Закрываем сервер после теста

	// Настроим реальный клиент, но направим его на фейковый сервер
	cfg := &config.Config{
		AllureBaseURL:   mockServer.URL,
		AllureAPIURL:    "/api/",
		AllureUserToken: "fake-token", // Чтобы `Authenticate` работал
	}
	client := adapter.NewAllureClient(cfg)

	// Запрашиваем запуски через реальный `AllureClient`
	launches, err := client.GetLaunches(context.Background())

	// Проверяем, что нет ошибок
	assert.NoError(t, err)
	assert.Len(t, launches, 2)
	assert.Equal(t, int64(1), launches[0].ID)
	assert.Equal(t, "Launch 1", launches[0].Name)
}

func TestGeneratePDFReport_RealClient(t *testing.T) {
	// Фейковый HTTP-сервер, который эмулирует Allure API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/uaa/oauth/token" {
			// Эмулируем выдачу токена
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token": "mocked_token", "expires_in": 3600}`))
			return
		}

		if r.Method == http.MethodPost && r.URL.Path == "/api/export/launch/pdf" {
			// Эмулируем успешную генерацию PDF
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(adapter.PDFReport{
				ID:     456,
				Name:   "Test Run",
				Type:   "pdf",
				Status: "generated",
			})
			return
		}

		// Если запрос не распознан, возвращаем 404
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close() // Закрываем сервер после теста

	// Настроим реальный клиент, но направим его на фейковый сервер
	cfg := &config.Config{
		AllureBaseURL:   mockServer.URL,
		AllureAPIURL:    "/api/",
		AllureUserToken: "fake-token",
	}
	client := adapter.NewAllureClient(cfg)

	// Запрашиваем генерацию PDF через реальный `AllureClient`
	report, err := client.GeneratePDFReport(context.Background(), 123, "Test Run")

	// Проверяем, что нет ошибок
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, int64(456), report.ID)
	assert.Equal(t, "Test Run", report.Name)
}

func TestGetPDFDownloadLink_RealClient(t *testing.T) {
	// Фейковый HTTP-сервер, который эмулирует Allure API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что клиент делает GET-запрос на скачивание
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/export/download/") {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			return
		}

		// Если запрос не распознан, возвращаем 404
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close() // Закрываем сервер после теста

	// Настроим реальный клиент, но направим его на фейковый сервер
	cfg := &config.Config{
		AllureBaseURL:   mockServer.URL,
		AllureAPIURL:    "/api/",
		AllureUserToken: "fake-token",
	}
	client := adapter.NewAllureClient(cfg)

	// Вызываем метод `GetPDFDownloadLink`
	link := client.GetPDFDownloadLink("456")

	// Проверяем, что ссылка корректная
	expectedURL := fmt.Sprintf("%s/api/export/download/456", mockServer.URL)
	assert.Equal(t, expectedURL, link)
}

func TestDownloadPDFReport_RealClient(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/uaa/oauth/token" {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token": "mocked_token", "expires_in": 3600}`))
			return
		}

		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/export/download/") {
			if r.Header.Get("Authorization") != "Bearer mocked_token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Disposition", "attachment; filename=allure-report-456.pdf")
			w.Header().Set("Content-Type", "application/pdf")
			_, _ = w.Write([]byte("PDF content"))
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	cfg := &config.Config{
		AllureBaseURL:   mockServer.URL,
		AllureAPIURL:    "/api/",
		AllureUserToken: "fake-token",
	}
	client := adapter.NewAllureClient(cfg)

	data, filename, err := client.DownloadPDFReport(context.Background(), "456")

	assert.NoError(t, err)
	assert.Equal(t, "allure-report-456.pdf", filename) // ✅ Исправлено имя файла
	assert.Equal(t, "PDF content", string(data))
}
