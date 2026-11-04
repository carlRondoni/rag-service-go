//go:build unit || !integration

package application_test

import (
	"context"
	"errors"
	"testing"

	"rag-service-go/internal/application"
	"rag-service-go/internal/domain"

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

type mockLLMClientForIngest struct {
	mock.Mock
}

func (m *mockLLMClientForIngest) Generate(ctx context.Context, prompt string) (string, error) {
	panic("not used")
}

func (m *mockLLMClientForIngest) Health(ctx context.Context) error {
	panic("not used")
}

func (m *mockLLMClientForIngest) Embed(ctx context.Context, texts []string) ([]domain.Embedding, error) {
	args := m.Called(ctx, texts)
	return args.Get(0).([]domain.Embedding), args.Error(1)
}

func TestIngestHandler_Handle_Success(t *testing.T) {
	ctx := context.Background()
	text := "Clean Architecture in Go"
	docID := "doc-123"

	chunks := []domain.Chunk{
		{ID: "c1", DocumentID: docID, Text: "Clean Architecture"},
		{ID: "c2", DocumentID: docID, Text: "in Go"},
	}
	embs := []domain.Embedding{
		{Vector: []float32{0.1, 0.2}},
		{Vector: []float32{0.3, 0.4}},
	}

	chunkerMock := new(mockChunker)
	chunkerMock.On("Chunk", text, docID).Return(chunks, nil)

	llmMock := new(mockLLMClientForIngest)
	llmMock.On("Embed", ctx, []string{"Clean Architecture", "in Go"}).Return(embs, nil)

	storeMock := new(mockVectorStore)
	storeMock.On("UpsertChunks", ctx, chunks, embs).Return(nil)

	handler := application.NewIngestHandler(chunkerMock, storeMock, llmMock)

	err := handler.Handle(ctx, text, docID)
	assert.NoError(t, err)

	chunkerMock.AssertExpectations(t)
	llmMock.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}

func TestIngestHandler_Handle_EmptyChunks(t *testing.T) {
	ctx := context.Background()
	text := ""
	docID := "doc-123"

	chunkerMock := new(mockChunker)
	chunkerMock.On("Chunk", text, docID).Return([]domain.Chunk{}, nil)

	storeMock := new(mockVectorStore)
	llmMock := new(mockLLMClientForIngest)

	handler := application.NewIngestHandler(chunkerMock, storeMock, llmMock)

	err := handler.Handle(ctx, text, docID)
	assert.NoError(t, err)

	chunkerMock.AssertExpectations(t)
}

func TestIngestHandler_Handle_ChunkerError(t *testing.T) {
	ctx := context.Background()
	text := "invalid"
	docID := "doc-123"
	expectedErr := errors.New("chunker failed")

	chunkerMock := new(mockChunker)
	chunkerMock.On("Chunk", text, docID).Return([]domain.Chunk{}, expectedErr)

	storeMock := new(mockVectorStore)
	llmMock := new(mockLLMClientForIngest)

	handler := application.NewIngestHandler(chunkerMock, storeMock, llmMock)

	err := handler.Handle(ctx, text, docID)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	chunkerMock.AssertExpectations(t)
}

func TestIngestHandler_Handle_EmbedError(t *testing.T) {
	ctx := context.Background()
	text := "some text"
	docID := "doc-123"
	expectedErr := errors.New("embed failed")

	chunks := []domain.Chunk{
		{ID: "c1", DocumentID: docID, Text: "some text"},
	}

	chunkerMock := new(mockChunker)
	chunkerMock.On("Chunk", text, docID).Return(chunks, nil)

	llmMock := new(mockLLMClientForIngest)
	llmMock.On("Embed", ctx, []string{"some text"}).Return([]domain.Embedding{}, expectedErr)

	storeMock := new(mockVectorStore)

	handler := application.NewIngestHandler(chunkerMock, storeMock, llmMock)

	err := handler.Handle(ctx, text, docID)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	chunkerMock.AssertExpectations(t)
	llmMock.AssertExpectations(t)
}

func TestIngestHandler_Handle_UpsertError(t *testing.T) {
	ctx := context.Background()
	text := "some text"
	docID := "doc-123"
	expectedErr := errors.New("upsert failed")

	chunks := []domain.Chunk{
		{ID: "c1", DocumentID: docID, Text: "some text"},
	}
	embs := []domain.Embedding{
		{Vector: []float32{0.1, 0.2}},
	}

	chunkerMock := new(mockChunker)
	chunkerMock.On("Chunk", text, docID).Return(chunks, nil)

	llmMock := new(mockLLMClientForIngest)
	llmMock.On("Embed", ctx, []string{"some text"}).Return(embs, nil)

	storeMock := new(mockVectorStore)
	storeMock.On("UpsertChunks", ctx, chunks, embs).Return(expectedErr)

	handler := application.NewIngestHandler(chunkerMock, storeMock, llmMock)

	err := handler.Handle(ctx, text, docID)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)

	chunkerMock.AssertExpectations(t)
	llmMock.AssertExpectations(t)
	storeMock.AssertExpectations(t)
}
