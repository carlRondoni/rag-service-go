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
	"github.com/stretchr/testify/mock"
)

type mockVectorStore struct {
	mock.Mock
}

func (m *mockVectorStore) UpsertChunks(ctx context.Context, chunks []domain.Chunk, embeddings []domain.Embedding) error {
	args := m.Called(ctx, chunks, embeddings)
	return args.Error(0)
}

func (m *mockVectorStore) Search(ctx context.Context, query domain.Embedding, topK int) ([]domain.SearchResult, error) {
	args := m.Called(ctx, query, topK)
	return args.Get(0).([]domain.SearchResult), args.Error(1)
}

func (m *mockVectorStore) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestDbHealthController_Success(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	store := &mockVectorStore{}
	store.On("Health", context.Background()).Return(nil)
	handler := application.NewDbHealthHandler(store)
	controller := controllers.NewDbHealthController(handler, logger)

	req := httptest.NewRequest(http.MethodGet, "/health/db", nil)
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var respBody map[string]string
	err := json.NewDecoder(res.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "ok", respBody["status"])
	assert.Equal(t, "up", respBody["db"])
}
