package application_test

import (
	"context"
	"rag-service-go/internal/application"
	"rag-service-go/internal/domain"
	"testing"

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

func TestDbHealthHandler_Handle_Success(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockVectorStore)
	mockStore.On("Health", ctx).Return(nil)

	handler := application.NewDbHealthHandler(mockStore)

	err := handler.Handle(ctx)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestDbHealthHandler_Handle_Error(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockVectorStore)
	mockStore.On("Health", ctx).Return(assert.AnError)

	handler := application.NewDbHealthHandler(mockStore)

	err := handler.Handle(ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	mockStore.AssertExpectations(t)
}
