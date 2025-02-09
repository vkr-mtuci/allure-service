package test

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/vkr-mtuci/allure-service/internal/adapter"
)

// MockAllureClient - мок-клиент Allure API
type MockAllureClient struct {
	mock.Mock
}

func (m *MockAllureClient) Authenticate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAllureClient) GetLaunches(ctx context.Context) ([]adapter.Launch, error) {
	args := m.Called(ctx)
	if launches, ok := args.Get(0).([]adapter.Launch); ok {
		return launches, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAllureClient) GeneratePDFReport(ctx context.Context, launchID int64, launchName string) (*adapter.PDFReport, error) {
	args := m.Called(ctx, launchID, launchName)
	if report, ok := args.Get(0).(*adapter.PDFReport); ok {
		return report, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAllureClient) GetPDFDownloadLink(reportID string) string {
	args := m.Called(reportID)
	return args.String(0)
}

func (m *MockAllureClient) DownloadPDFReport(ctx context.Context, reportID string) ([]byte, string, error) {
	args := m.Called(ctx, reportID)
	if data, ok := args.Get(0).([]byte); ok {
		return data, args.String(1), args.Error(2)
	}
	return nil, "", args.Error(2)
}

// MockAllureService - мок-сервис для AllureService
type MockAllureService struct {
	mock.Mock
}

// GetNextLaunch - мок-метод поиска ближайшего запуска
func (m *MockAllureService) GetNextLaunch(afterDate time.Time) (*adapter.Launch, error) {
	args := m.Called(afterDate)
	if launch, ok := args.Get(0).(*adapter.Launch); ok {
		return launch, args.Error(1)
	}
	return nil, args.Error(1)
}

// GeneratePDFReport - мок-метод генерации PDF
func (m *MockAllureService) GeneratePDFReport(launchID int64, launchName string) (*adapter.PDFReport, error) {
	args := m.Called(launchID, launchName)
	if report, ok := args.Get(0).(*adapter.PDFReport); ok {
		return report, args.Error(1)
	}
	return nil, args.Error(1)
}

// GetPDFDownloadLink - мок-метод получения ссылки PDF
func (m *MockAllureService) GetPDFDownloadLink(reportID string) string {
	args := m.Called(reportID)
	return args.String(0)
}

// DownloadPDFReport - мок-метод скачивания PDF
func (m *MockAllureService) DownloadPDFReport(reportID string) ([]byte, string, error) {
	args := m.Called(reportID)
	if fileData, ok := args.Get(0).([]byte); ok {
		return fileData, args.String(1), args.Error(2)
	}
	return nil, "", args.Error(2)
}
