//go:build unit || !integration

package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"rag-service-go/internal/application"
	"rag-service-go/internal/domain"

	"github.com/stretchr/testify/assert"
)

const keyGenerateConst ctxKey = "key"

type mockLLMGenerateTest struct {
	generateFn func(ctx context.Context, prompt string) (string, error)
	embedFn    func(ctx context.Context, texts []string) ([]domain.Embedding, error)
}

func (m mockLLMGenerateTest) Generate(ctx context.Context, prompt string) (string, error) {
	if m.generateFn != nil {
		return m.generateFn(ctx, prompt)
	}
	return "", nil
}

func (m mockLLMGenerateTest) Stream(ctx context.Context, prompt string) (<-chan string, error) {
	panic("not used")
}

func (m mockLLMGenerateTest) Health(ctx context.Context) error {
	panic("not used")
}

func (m mockLLMGenerateTest) Embed(ctx context.Context, texts []string) ([]domain.Embedding, error) {
	if m.embedFn != nil {
		return m.embedFn(ctx, texts)
	}
	return []domain.Embedding{{Vector: []float32{0.1, 0.2}}}, nil
}

func TestGenerateHandler_Success_WithContext(t *testing.T) {
	expected := "Clean Architecture is great"
	queryPrompt := "Tell me about architecture"

	mockLLM := mockLLMGenerateTest{
		embedFn: func(ctx context.Context, texts []string) ([]domain.Embedding, error) {
			assert.Equal(t, []string{queryPrompt}, texts)
			return []domain.Embedding{{Vector: []float32{0.1, 0.2}}}, nil
		},
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			assert.True(t, strings.Contains(prompt, "Clean Architecture is a software design philosophy"))
			assert.True(t, strings.Contains(prompt, queryPrompt))
			return expected, nil
		},
	}

	mockStore := new(mockVectorStore)
	mockStore.On("Search", context.Background(), domain.Embedding{Vector: []float32{0.1, 0.2}}, 3).
		Return([]domain.SearchResult{
			{
				Score: 0.95,
				Chunk: domain.Chunk{
					ID:         "chunk-1",
					DocumentID: "doc-1",
					Text:       "Clean Architecture is a software design philosophy",
				},
			},
		}, nil)

	h := application.NewGenerateHandler(mockLLM, mockStore)

	res, err := h.Handle(context.Background(), queryPrompt)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	mockStore.AssertExpectations(t)
}

func TestGenerateHandler_Success_NoContextFound(t *testing.T) {
	expected := "direct answer without context"
	queryPrompt := "general question"

	mockLLM := mockLLMGenerateTest{
		embedFn: func(ctx context.Context, texts []string) ([]domain.Embedding, error) {
			return []domain.Embedding{{Vector: []float32{0.1, 0.2}}}, nil
		},
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			assert.Equal(t, queryPrompt, prompt)
			return expected, nil
		},
	}

	mockStore := new(mockVectorStore)
	mockStore.On("Search", context.Background(), domain.Embedding{Vector: []float32{0.1, 0.2}}, 3).
		Return([]domain.SearchResult{}, nil)

	h := application.NewGenerateHandler(mockLLM, mockStore)

	res, err := h.Handle(context.Background(), queryPrompt)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	mockStore.AssertExpectations(t)
}

func TestGenerateHandler_EmbedError(t *testing.T) {
	expectedErr := errors.New("embed failed")

	mockLLM := mockLLMGenerateTest{
		embedFn: func(ctx context.Context, texts []string) ([]domain.Embedding, error) {
			return nil, expectedErr
		},
	}

	mockStore := new(mockVectorStore)

	h := application.NewGenerateHandler(mockLLM, mockStore)

	_, err := h.Handle(context.Background(), "prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to embed query prompt")
}

func TestGenerateHandler_SearchError(t *testing.T) {
	expectedErr := errors.New("search failed")

	mockLLM := mockLLMGenerateTest{
		embedFn: func(ctx context.Context, texts []string) ([]domain.Embedding, error) {
			return []domain.Embedding{{Vector: []float32{0.1, 0.2}}}, nil
		},
	}

	mockStore := new(mockVectorStore)
	mockStore.On("Search", context.Background(), domain.Embedding{Vector: []float32{0.1, 0.2}}, 3).
		Return([]domain.SearchResult{}, expectedErr)

	h := application.NewGenerateHandler(mockLLM, mockStore)

	_, err := h.Handle(context.Background(), "prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to search vector store")
	mockStore.AssertExpectations(t)
}

func TestGenerateHandler_LLMGenerateError(t *testing.T) {
	expectedErr := errors.New("llm failed")

	mockLLM := mockLLMGenerateTest{
		embedFn: func(ctx context.Context, texts []string) ([]domain.Embedding, error) {
			return []domain.Embedding{{Vector: []float32{0.1, 0.2}}}, nil
		},
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return "", expectedErr
		},
	}

	mockStore := new(mockVectorStore)
	mockStore.On("Search", context.Background(), domain.Embedding{Vector: []float32{0.1, 0.2}}, 3).
		Return([]domain.SearchResult{}, nil)

	h := application.NewGenerateHandler(mockLLM, mockStore)

	_, err := h.Handle(context.Background(), "prompt")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockStore.AssertExpectations(t)
}

func TestGenerateHandler_ContextForwarded(t *testing.T) {
	mockLLM := mockLLMGenerateTest{
		embedFn: func(ctx context.Context, texts []string) ([]domain.Embedding, error) {
			assert.Equal(t, "value", ctx.Value(keyGenerateConst))
			return []domain.Embedding{{Vector: []float32{0.1, 0.2}}}, nil
		},
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			assert.Equal(t, "value", ctx.Value(keyGenerateConst))
			return "ok", nil
		},
	}

	mockStore := new(mockVectorStore)
	ctx := context.WithValue(context.Background(), keyGenerateConst, "value")
	mockStore.On("Search", ctx, domain.Embedding{Vector: []float32{0.1, 0.2}}, 3).
		Return([]domain.SearchResult{}, nil)

	h := application.NewGenerateHandler(mockLLM, mockStore)

	_, err := h.Handle(ctx, "hi")
	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}
