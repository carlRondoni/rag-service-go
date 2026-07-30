package application_test

import (
	"context"
	"errors"
	"testing"

	"rag-service-go/internal/application"
	"rag-service-go/internal/domain"

	"github.com/stretchr/testify/assert"
)

const keyLLmHealthCheckConst string = "key"

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

func TestLLMHealthCheckHandler_CallsClient(t *testing.T) {
	mock := &mockLLMHealthCheck{}
	h := application.NewLLMHealthCheckHandler(mock)

	ctx := context.WithValue(context.Background(), keyLLmHealthCheckConst, "value")

	err := h.Handle(ctx)
	assert.NoError(t, err)
	assert.True(t, mock.called)
	assert.Equal(t, "value", mock.ctx.Value(keyLLmHealthCheckConst))
}

func TestLLMHealthCheckHandler_ReturnsError(t *testing.T) {
	expectedErr := errors.New("boom")

	mock := &mockLLMHealthCheck{
		err: expectedErr,
	}

	h := application.NewLLMHealthCheckHandler(mock)

	err := h.Handle(context.Background())
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}
