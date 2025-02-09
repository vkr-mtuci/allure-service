package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vkr-mtuci/allure-service/config"
)

// Config Test
func TestLoadConfig(t *testing.T) {
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("ALLURE_BASE_URL", "https://allure.example.com")
	os.Setenv("ALLURE_API_URL", "/api")
	os.Setenv("ALLURE_API_TOKEN", "test-token")
	os.Setenv("ALLURE_PROJECT_ID", "1661")

	cfg := config.LoadConfig()

	assert.Equal(t, "8080", cfg.ServerPort)
	assert.Equal(t, "https://allure.example.com", cfg.AllureBaseURL)
	assert.Equal(t, "/api", cfg.AllureAPIURL)
	assert.Equal(t, "test-token", cfg.AllureUserToken)
	assert.Equal(t, "1661", cfg.AllureProjectID)
}
