//go:build integration

package vector_store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"rag-service-go/internal/domain"
	"rag-service-go/internal/infrastructure/vector_store"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getQdrantURL() string {
	if url := os.Getenv("QDRANT_URL"); url != "" {
		return url
	}
	return "http://localhost:6333"
}

func TestQdrantStore_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	baseURL := getQdrantURL()
	logger := zerolog.New(os.Stdout)
	collection := "integration_test_collection"

	store := vector_store.NewQdrantStore(baseURL, collection, logger)

	err := store.Health(ctx)
	require.NoError(t, err, "Qdrant must be running at %s for integration tests (start via make docker-start)", baseURL)

	err = store.EnsureCollection(ctx, 3)
	require.NoError(t, err)

	chunks := []domain.Chunk{
		{
			ID:         "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			DocumentID: "doc-integration-1",
			Text:       "Go is built for concurrency and simplicity.",
		},
		{
			ID:         "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22",
			DocumentID: "doc-integration-2",
			Text:       "Python is widely used in data science.",
		},
	}
	embeddings := []domain.Embedding{
		{Vector: []float32{1.0, 0.0, 0.0}},
		{Vector: []float32{0.0, 1.0, 0.0}},
	}

	err = store.UpsertChunks(ctx, chunks, embeddings)
	require.NoError(t, err)

	// let qdrant index the points
	time.Sleep(500 * time.Millisecond)

	query := domain.Embedding{Vector: []float32{0.9, 0.1, 0.0}}
	results, err := store.Search(ctx, query, 2)
	require.NoError(t, err)
	require.NotEmpty(t, results)

	assert.Equal(t, "doc-integration-1", results[0].Chunk.DocumentID)
	assert.Equal(t, "Go is built for concurrency and simplicity.", results[0].Chunk.Text)
	assert.Greater(t, results[0].Score, float32(0.8))
}

