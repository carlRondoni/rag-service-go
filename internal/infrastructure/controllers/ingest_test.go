//go:build unit || !integration

package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

type mockChunkerForController struct {
	mock.Mock
}

func (m *mockChunkerForController) Chunk(text string, documentID string) ([]domain.Chunk, error) {
	args := m.Called(text, documentID)
	return args.Get(0).([]domain.Chunk), args.Error(1)
}

type mockLLMForController struct {
	mock.Mock
}

func (m *mockLLMForController) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *mockLLMForController) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockLLMForController) Embed(ctx context.Context, texts []string) ([]domain.Embedding, error) {
	args := m.Called(ctx, texts)
	return args.Get(0).([]domain.Embedding), args.Error(1)
}

func TestIngestController_Success(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	chunkerMock := new(mockChunkerForController)
	storeMock := new(mockVectorStore)
	llmMock := new(mockLLMForController)

	chunks := []domain.Chunk{{ID: "c1", DocumentID: "doc-1", Text: "sample text"}}
	embs := []domain.Embedding{{Vector: []float32{0.1, 0.2}}}

	chunkerMock.On("Chunk", "sample text", "doc-1").Return(chunks, nil)
	llmMock.On("Embed", mock.Anything, []string{"sample text"}).Return(embs, nil)
	storeMock.On("UpsertChunks", mock.Anything, chunks, embs).Return(nil)

	handler := application.NewIngestHandler(chunkerMock, storeMock, llmMock)
	controller := controllers.NewIngestController(handler, logger)

	body := map[string]string{
		"text":       "sample text",
		"documentId": "doc-1",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer(b))
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer func() { assert.NoError(t, res.Body.Close()) }()

	assert.Equal(t, http.StatusAccepted, res.StatusCode)
}

func TestIngestController_MethodNotAllowed(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	controller := controllers.NewIngestController(application.IngestHandler{}, logger)

	req := httptest.NewRequest(http.MethodGet, "/ingest", nil)
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer func() { assert.NoError(t, res.Body.Close()) }()

	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestIngestController_MissingText(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	controller := controllers.NewIngestController(application.IngestHandler{}, logger)

	body := map[string]string{
		"documentId": "doc-1",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer(b))
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer func() { assert.NoError(t, res.Body.Close()) }()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestIngestController_HandlerError(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	chunkerMock := new(mockChunkerForController)
	storeMock := new(mockVectorStore)
	llmMock := new(mockLLMForController)

	chunkerMock.On("Chunk", "sample text", "default-doc").Return([]domain.Chunk{}, errors.New("chunker boom"))

	handler := application.NewIngestHandler(chunkerMock, storeMock, llmMock)
	controller := controllers.NewIngestController(handler, logger)

	body := map[string]string{
		"text": "sample text",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer(b))
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer func() { assert.NoError(t, res.Body.Close()) }()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

