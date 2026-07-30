package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"rag-service-go/internal/application"
	"rag-service-go/internal/domain"
	"rag-service-go/internal/infrastructure/controllers"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type mockLLMHealthCheck struct {
	called bool
	ctx    context.Context
	err    error
}

func (m *mockLLMHealthCheck) Health(ctx context.Context) error {
	m.called = true
	m.ctx = ctx
	return m.err
}

func (m *mockLLMHealthCheck) Generate(ctx context.Context, prompt string) (string, error) {
	panic("not used")
}

func (m *mockLLMHealthCheck) Stream(ctx context.Context, prompt string) (<-chan string, error) {
	panic("not used")
}

func (m *mockLLMHealthCheck) Embed(ctx context.Context, texts []string) ([]domain.Embedding, error) {
	panic("not used")
}

func TestLlmHealthCheckController_Success(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	client := &mockLLMHealthCheck{}
	handler := application.NewLLMHealthCheckHandler(client)
	controller := controllers.NewLlmHealthCheckController(handler, logger)

	req := httptest.NewRequest(http.MethodGet, "/health/llm", nil)
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(res.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "ok", respBody["status"])
	assert.Equal(t, "up", respBody["llm"])
}
