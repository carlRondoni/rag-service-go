package domain

import "context"

type Embedding struct {
	Vector []float32
}

type SearchResult struct {
	Chunk Chunk
	Score float32
}

type VectorStore interface {
	UpsertChunks(ctx context.Context, chunks []Chunk, embeddings []Embedding) error
	Search(ctx context.Context, query Embedding, topK int) ([]SearchResult, error)
}
