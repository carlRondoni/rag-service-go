package application_test

import (
	"context"
	"errors"
	"rag-service-go/internal/application"
	"rag-service-go/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockLLMGenerateTest struct {
	generateFn func(ctx context.Context, prompt string) (string, error)
}

func (m mockLLMGenerateTest) Generate(ctx context.Context, prompt string) (string, error) {
	return m.generateFn(ctx, prompt)
}

func (m mockLLMGenerateTest) Stream(ctx context.Context, prompt string) (<-chan string, error) {
	panic("not used")
}

func (m mockLLMGenerateTest) Health(ctx context.Context) error {
	panic("not used")
}

func (m mockLLMGenerateTest) Embed(ctx context.Context, texts []string) ([]domain.Embedding, error) {
	panic("not used")
}

func TestGenerateHandler_Success(t *testing.T) {
	expected := "hello world"

	mock := mockLLMGenerateTest{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			if prompt != "hi" {
				t.Fatalf("unexpected prompt: %s", prompt)
			}
			return expected, nil
		},
	}

	h := application.NewGenerateHandler(mock)

	res, err := h.Handle(context.Background(), "hi")
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestGenerateHandler_ErrorPropagation(t *testing.T) {
	expectedErr := errors.New("llm failed")

	mock := mockLLMGenerateTest{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			return "", expectedErr
		},
	}

	h := application.NewGenerateHandler(mock)

	_, err := h.Handle(context.Background(), "hi")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestGenerateHandler_ContextForwarded(t *testing.T) {
	mock := mockLLMGenerateTest{
		generateFn: func(ctx context.Context, prompt string) (string, error) {
			assert.Equal(t, "value", ctx.Value("key"))
			return "ok", nil
		},
	}

	h := application.NewGenerateHandler(mock)

	ctx := context.WithValue(context.Background(), "key", "value")

	_, _ = h.Handle(ctx, "hi")
}
