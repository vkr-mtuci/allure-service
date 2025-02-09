package test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vkr-mtuci/allure-service/internal/adapter"
	"github.com/vkr-mtuci/allure-service/internal/service"
)

// ✅ **Тест: Успешный поиск ближайшего запуска**
func TestGetNextLaunch_Success(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)

	now := time.Now().UnixMilli()
	mockLaunches := []adapter.Launch{
		{ID: 101, Name: "Older Run", CreatedDate: now - 50000},
		{ID: 102, Name: "Latest Run", CreatedDate: now + 1000}, // Ближайший запуск после переданной даты
	}

	mockClient.On("GetLaunches", mock.Anything).Return(mockLaunches, nil)

	launch, err := service.GetNextLaunch(time.Now()) // Передаем текущее время, а не -2 часа
	assert.NoError(t, err)
	assert.NotNil(t, launch)
	assert.Equal(t, int64(102), launch.ID) // Теперь этот запуск действительно ближайший
}

// ❌ **Тест: Нет подходящих запусков**
func TestGetNextLaunch_NoLaunches(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)

	mockClient.On("GetLaunches", mock.Anything).Return([]adapter.Launch{}, nil)

	launch, err := service.GetNextLaunch(time.Now().Add(-1 * time.Hour))
	assert.Error(t, err)
	assert.Nil(t, launch)
}

// 🚨 **Тест: Ошибка API**
func TestGetNextLaunch_APIError(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)

	// ✅ Возвращаем **пустой** слайс `[]adapter.Launch{}` вместо `nil`
	mockClient.On("GetLaunches", mock.Anything).Return([]adapter.Launch{}, errors.New("ошибка API"))

	launch, err := service.GetNextLaunch(time.Now().Add(-1 * time.Hour))
	assert.Error(t, err)
	assert.Nil(t, launch)
}

// ✅ **Тест: Успешная генерация PDF-отчёта**
func TestGeneratePDFReport_Success(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)

	mockReport := &adapter.PDFReport{
		ID:          999,
		Name:        "Test Report",
		ProjectID:   1661,
		CreatedDate: time.Now().UnixMilli(),
	}

	mockClient.On("GeneratePDFReport", mock.Anything, int64(123), "Test Run").Return(mockReport, nil)

	report, err := service.GeneratePDFReport(123, "Test Run")
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, int64(999), report.ID)
}

// ❌ **Тест: Ошибка генерации PDF**
func TestGeneratePDFReport_Error(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)

	mockClient.On("GeneratePDFReport", mock.Anything, int64(123), "Test Run").
		Return((*adapter.PDFReport)(nil), errors.New("ошибка генерации PDF"))

	report, err := service.GeneratePDFReport(123, "Test Run")
	assert.Error(t, err)
	assert.Nil(t, report)
}

// ✅ **Тест: Успешное скачивание PDF**
func TestDownloadPDFReport_Success(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)

	pdfContent := []byte("PDF FILE CONTENT")
	fileName := "allure-report-999.pdf"

	mockClient.On("DownloadPDFReport", mock.Anything, "999").Return(pdfContent, fileName, nil)

	data, name, err := service.DownloadPDFReport("999")
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, fileName, name)
}

// ❌ **Тест: Ошибка скачивания PDF**
func TestDownloadPDFReport_Error(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)

	mockClient.On("DownloadPDFReport", mock.Anything, "999").
		Return(([]byte)(nil), "", errors.New("ошибка скачивания PDF")) // ✅ Теперь безопасно

	data, name, err := service.DownloadPDFReport("999")
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Equal(t, "", name)
}

func TestGeneratePDFReport_InvalidParameters(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)

	// Добавляем мокирование для вызовов с некорректными параметрами
	mockClient.On("GeneratePDFReport", mock.Anything, int64(0), "Test").
		Return(nil, errors.New("invalid launch ID"))

	mockClient.On("GeneratePDFReport", mock.Anything, int64(123), "").
		Return(nil, errors.New("empty launch name"))

	// Тест с нулевым LaunchID
	_, err := service.GeneratePDFReport(0, "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid launch ID")

	// Тест с пустым именем
	_, err = service.GeneratePDFReport(123, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty launch name")

	// Проверяем, что мокированные вызовы действительно произошли
	mockClient.AssertExpectations(t)
}

func TestDownloadPDFReport_EdgeCases(t *testing.T) {
	mockClient := new(MockAllureClient)
	service := service.NewAllureService(mockClient)

	// Добавляем мокирование вызова с пустым reportID
	mockClient.On("DownloadPDFReport", mock.Anything, "").Return(
		nil, "", errors.New("empty report ID"),
	)

	// Добавляем мокирование вызова с неверным форматом ID
	mockClient.On("DownloadPDFReport", mock.Anything, "invalid").Return(
		nil, "", errors.New("invalid report ID"),
	)

	// Тест с пустым reportID
	_, _, err := service.DownloadPDFReport("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty report ID")

	// Тест с неверным форматом ID
	_, _, err = service.DownloadPDFReport("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid report ID")

	// Проверяем, что моки были вызваны
	mockClient.AssertExpectations(t)
}
