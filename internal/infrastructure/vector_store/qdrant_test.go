//go:build unit || !integration

package vector_store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"rag-service-go/internal/domain"
	"rag-service-go/internal/infrastructure/vector_store"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestQdrantStore_Health(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	logger := zerolog.New(os.Stdout)
	store := vector_store.NewQdrantStore(srv.URL, "test-collection", logger)

	err := store.Health(context.Background())
	assert.NoError(t, err)
}

func TestQdrantStore_EnsureCollection_CreatesWhenNotFound(t *testing.T) {
	created := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/test-collection" {
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if r.Method == http.MethodPut {
				created = true
				w.WriteHeader(http.StatusOK)
				_, _ = fmt.Fprint(w, `{"result": true, "status": "ok"}`)
				return
			}
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	logger := zerolog.New(os.Stdout)
	store := vector_store.NewQdrantStore(srv.URL, "test-collection", logger)

	err := store.EnsureCollection(context.Background(), 768)
	assert.NoError(t, err)
	assert.True(t, created)
}

func TestQdrantStore_UpsertChunks_Success(t *testing.T) {
	upsertCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/test-collection" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/collections/test-collection/points" && r.Method == http.MethodPut {
			upsertCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"result": {"operation_id": 1, "status": "completed"}, "status": "ok"}`)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	logger := zerolog.New(os.Stdout)
	store := vector_store.NewQdrantStore(srv.URL, "test-collection", logger)

	chunks := []domain.Chunk{
		{ID: "550e8400-e29b-41d4-a716-446655440000", DocumentID: "doc-1", Text: "Hello world"},
	}
	embeddings := []domain.Embedding{
		{Vector: []float32{0.1, 0.2, 0.3}},
	}

	err := store.UpsertChunks(context.Background(), chunks, embeddings)
	assert.NoError(t, err)
	assert.True(t, upsertCalled)
}

func TestQdrantStore_Search_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/test-collection/points/search" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{
				"result": [
					{
						"id": "550e8400-e29b-41d4-a716-446655440000",
						"score": 0.89,
						"payload": {
							"text": "Hello world",
							"document_id": "doc-1"
						}
					}
				]
			}`)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	logger := zerolog.New(os.Stdout)
	store := vector_store.NewQdrantStore(srv.URL, "test-collection", logger)

	results, err := store.Search(context.Background(), domain.Embedding{Vector: []float32{0.1, 0.2}}, 3)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, float32(0.89), results[0].Score)
	assert.Equal(t, "Hello world", results[0].Chunk.Text)
	assert.Equal(t, "doc-1", results[0].Chunk.DocumentID)
}

func TestQdrantStore_Search_RequestsPayload(t *testing.T) {
	var body map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		_, _ = fmt.Fprint(w, `{"result":[]}`)
	}))
	defer srv.Close()

	store := vector_store.NewQdrantStore(srv.URL, "test-collection", zerolog.Nop())

	_, err := store.Search(context.Background(), domain.Embedding{Vector: []float32{0.1}}, 3)
	assert.NoError(t, err)
	assert.Equal(t, true, body["with_payload"])
}

func TestQdrantStore_Search_NotFoundReturnsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	logger := zerolog.New(os.Stdout)
	store := vector_store.NewQdrantStore(srv.URL, "test-collection", logger)

	results, err := store.Search(context.Background(), domain.Embedding{Vector: []float32{0.1, 0.2}}, 3)
	assert.NoError(t, err)
	assert.Empty(t, results)
}
