package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"rag-service-go/internal/infrastructure/controllers"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckController_Integration(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	controller := controllers.NewHealthCheckController(logger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}
