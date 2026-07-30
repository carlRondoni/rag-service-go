package application_test

import (
	"context"
	"rag-service-go/internal/application"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIngestHandler_Handle_Success(t *testing.T) {
	ctx := context.Background()
	mockStore := new(mockVectorStore)
	mockStore.On("Health", ctx).Return(nil)

	handler := application.NewDbHealthHandler(mockStore)

	err := handler.Handle(ctx)
	assert.NoError(t, err)

	mockStore.AssertExpectations(t)
}
