package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/vkr-mtuci/allure-service/config"
)

// AllureClientInterface - интерфейс клиента Allure API
type AllureClientInterface interface {
	Authenticate(ctx context.Context) error
	GetLaunches(ctx context.Context) ([]Launch, error)
	GeneratePDFReport(ctx context.Context, launchID int64, launchName string) (*PDFReport, error)
	GetPDFDownloadLink(reportID string) string
	DownloadPDFReport(ctx context.Context, reportID string) ([]byte, string, error)
}

// AllureClient - клиент API Allure
type AllureClient struct {
	client       *resty.Client
	baseURL      string
	apiURL       string
	token        string
	tokenExpires time.Time
	projectID    string
	cfg          *config.Config
	mu           sync.Mutex // Добавляем мьютекс
}

// NewAllureClient - создание клиента API Allure
func NewAllureClient(cfg *config.Config) *AllureClient {
	client := resty.New().
		SetBaseURL(cfg.AllureBaseURL).
		SetTimeout(10*time.Second).
		SetHeader("Accept", "application/json")

	return &AllureClient{
		client:    client,
		baseURL:   cfg.AllureBaseURL,
		apiURL:    cfg.AllureAPIURL,
		token:     cfg.AllureUserToken,
		projectID: cfg.AllureProjectID,
		cfg:       cfg,
	}
}

// Authenticate - проверяет и обновляет токен, если он истек
func (a *AllureClient) Authenticate(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Если токен еще валиден, используем его
	if time.Until(a.tokenExpires) > 5*time.Minute {
		return nil
	}

	log.Info().Msg("🔄 Обновление токена Allure API...")

	// Отправляем запрос на обновление токена
	resp, err := a.client.R().
		SetContext(ctx).
		SetFormData(map[string]string{
			"grant_type": "apitoken",
			"scope":      "openid",
			"token":      a.cfg.AllureUserToken,
		}).
		Post(a.baseURL + "/api/uaa/oauth/token")

	if err != nil {
		log.Error().Err(err).Msg("❌ Ошибка обновления токена")
		return fmt.Errorf("ошибка обновления токена: %w", err)
	}

	// Логируем полный ответ от Allure API
	log.Info().Msgf("📨 Ответ от Allure API (токен): статус %d, тело: %s", resp.StatusCode(), resp.String())

	// Проверяем статус ответа
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("ошибка обновления токена: статус %d", resp.StatusCode())
	}

	// Парсим JSON-ответ
	var authResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(resp.Body(), &authResp); err != nil {
		log.Error().Err(err).Msg("❌ Ошибка парсинга токена")
		return err
	}

	// Сохраняем новый токен
	a.token = authResp.AccessToken
	a.tokenExpires = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)
	log.Info().Msg("✅ Токен успешно обновлен!")
	return nil
}

// GetLaunches - получает список запусков
func (a *AllureClient) GetLaunches(ctx context.Context) ([]Launch, error) {
	// Убеждаемся, что токен актуален
	if err := a.Authenticate(ctx); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%slaunch?projectId=%s&page=0&size=100", a.baseURL, a.apiURL, a.projectID)

	resp, err := a.client.R().
		SetContext(ctx).
		SetAuthToken(a.token).
		Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("ошибка Allure API: статус %d", resp.StatusCode())
	}

	var launchesResponse struct {
		Content []Launch `json:"content"`
	}

	if err := json.Unmarshal(resp.Body(), &launchesResponse); err != nil {
		return nil, err
	}

	return launchesResponse.Content, nil
}

// GeneratePDFReport - инициирует создание PDF-отчета в Allure
func (a *AllureClient) GeneratePDFReport(ctx context.Context, launchID int64, launchName string) (*PDFReport, error) {
	// Обновляем токен перед запросом
	err := a.Authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка авторизации перед генерацией PDF: %w", err)
	}

	url := fmt.Sprintf("%s%sexport/launch/pdf", a.baseURL, a.apiURL)
	log.Info().Msgf("📡 Отправка запроса на генерацию PDF: URL=%s, LaunchID=%d", url, launchID)

	// Формируем JSON-запрос
	requestBody := map[string]interface{}{
		"launchId":        launchID,
		"name":            launchName,
		"withPageNumbers": true,
	}

	// Отправляем запрос
	resp, err := a.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+a.token). // ✅ Добавляем Bearer-токен
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post(url)

	if err != nil {
		log.Error().Err(err).Msg("❌ Ошибка запроса на генерацию PDF")
		return nil, err
	}

	// Логируем полный ответ от Allure API
	log.Info().Msgf("📨 Ответ от Allure API: статус %d, тело: %s", resp.StatusCode(), resp.String())

	// Проверяем статус ответа
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("ошибка генерации PDF: статус %d", resp.StatusCode())
	}

	// Парсим JSON-ответ
	var report PDFReport
	if err := json.Unmarshal(resp.Body(), &report); err != nil {
		log.Error().Err(err).Msg("❌ Ошибка парсинга ответа на генерацию PDF")
		return nil, err
	}

	log.Info().Msgf("✅ Успешная генерация PDF: ID=%d, имя=%s", report.ID, report.Name)
	return &report, nil
}

// GetPDFDownloadLink - получает ссылку на скачивание PDF-отчета
func (a *AllureClient) GetPDFDownloadLink(reportID string) string {
	return fmt.Sprintf("%s%sexport/download/%s", a.baseURL, a.apiURL, reportID)
}

// DownloadPDFReport - загружает PDF-отчет с Allure API
func (a *AllureClient) DownloadPDFReport(ctx context.Context, reportID string) ([]byte, string, error) {
	// Обновляем токен перед скачиванием
	if err := a.Authenticate(ctx); err != nil {
		return nil, "", fmt.Errorf("❌ Ошибка авторизации перед скачиванием PDF: %w", err)
	}

	url := fmt.Sprintf("%s%sexport/download/%s", a.baseURL, a.apiURL, reportID)
	log.Info().Msgf("📡 Запрос на скачивание PDF: %s", url)

	resp, err := a.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+a.token).
		Get(url)

	if err != nil {
		log.Error().Err(err).Msg("❌ Ошибка при скачивании PDF")
		return nil, "", err
	}

	if resp.StatusCode() != http.StatusOK {
		log.Warn().Msgf("⚠️ Ошибка скачивания PDF: статус %d", resp.StatusCode())
		return nil, "", fmt.Errorf("ошибка скачивания PDF: статус %d", resp.StatusCode())
	}

	// Определяем имя файла
	fileName := fmt.Sprintf("allure-report-%s.pdf", reportID)

	return resp.Body(), fileName, nil
}
