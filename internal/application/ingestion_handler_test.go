package application_test

import (
	"context"
	"rag-service-go/internal/application"
	"rag-service-go/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockChunker struct {
	mock.Mock
}

func (m *mockChunker) Chunk(text string, documentID string) ([]domain.Chunk, error) {
	args := m.Called(text, documentID)
	return args.Get(0).([]domain.Chunk), args.Error(1)
}

type mockLLMClient struct {
	mock.Mock
}

func (m *mockLLMClient) Embed(ctx context.Context, texts []string) ([]domain.Embedding, error) {
	args := m.Called(ctx, texts)
	return args.Get(0).([]domain.Embedding), args.Error(1)
}

func (m *mockLLMClient) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestIngestHandler_Handle_Success(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockVectorStore)
	mockStore.On("Health", ctx).Return(nil)

	handler := application.NewDbHealthHandler(mockStore)

	err := handler.Handle(ctx)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}
