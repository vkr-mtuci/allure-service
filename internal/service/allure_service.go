package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vkr-mtuci/allure-service/internal/adapter"
)

// Интерфейс сервиса
type AllureServiceInterface interface {
	GetNextLaunch(afterDate time.Time) (*adapter.Launch, error)
	GeneratePDFReport(launchID int64, launchName string) (*adapter.PDFReport, error)
	GetPDFDownloadLink(reportID string) string
	DownloadPDFReport(reportID string) ([]byte, string, error)
}

// AllureService - реализация сервиса
type AllureService struct {
	client adapter.AllureClientInterface
}

// NewAllureService - создание сервиса
func NewAllureService(client adapter.AllureClientInterface) *AllureService {
	return &AllureService{client: client}
}

// GetNextLaunch - поиск ближайшего запуска после переданной даты
func (s *AllureService) GetNextLaunch(afterDate time.Time) (*adapter.Launch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	launches, err := s.client.GetLaunches(ctx)
	if err != nil {
		log.Error().Err(err).Msg("❌ Ошибка получения запусков")
		return nil, err
	}

	if len(launches) == 0 {
		log.Warn().Msg("⚠️ Нет запусков для поиска")
		return nil, fmt.Errorf("нет запусков для анализа")
	}

	afterTimestamp := afterDate.UnixMilli()
	var closestLaunch *adapter.Launch
	var minDiff int64 = math.MaxInt64

	for _, launch := range launches {
		diff := launch.CreatedDate - afterTimestamp
		if diff >= 0 && diff < minDiff {
			closestLaunch = &launch
			minDiff = diff
		}
	}

	if closestLaunch == nil {
		log.Warn().Msg("⚠️ Не найден запуск после указанной даты")
		return nil, fmt.Errorf("не найден запуск после указанной даты")
	}

	log.Info().Msgf("✅ Найден ближайший запуск: %s (ID: %d)", closestLaunch.Name, closestLaunch.ID)
	return closestLaunch, nil
}

// GeneratePDFReport - инициирует создание PDF-отчета
func (s *AllureService) GeneratePDFReport(launchID int64, launchName string) (*adapter.PDFReport, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	report, err := s.client.GeneratePDFReport(ctx, launchID, launchName)
	if err != nil {
		log.Error().Err(err).Msg("❌ Ошибка генерации PDF-отчета")
		return nil, err
	}

	log.Info().Msgf("✅ PDF-отчет сгенерирован: %s (ID: %d)", report.Name, report.ID)
	return report, nil
}

// GetPDFDownloadLink - формирует ссылку для скачивания PDF-отчета
func (s *AllureService) GetPDFDownloadLink(reportID string) string {
	return s.client.GetPDFDownloadLink(reportID)
}

// DownloadPDFReport - скачивает PDF-отчет и отдает его фронтенду
func (s *AllureService) DownloadPDFReport(reportID string) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fileData, fileName, err := s.client.DownloadPDFReport(ctx, reportID)
	if err != nil {
		log.Error().Err(err).Msg("❌ Ошибка скачивания PDF-отчета")
		return nil, "", err
	}

	log.Info().Msgf("✅ PDF-отчет скачан: %s", fileName)
	return fileData, fileName, nil
}
