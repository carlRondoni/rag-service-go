package application

import (
	"context"
	"rag-service-go/internal/domain"
)

type IngestHandler struct {
	chunker   domain.Chunker
	store     domain.VectorStore
	llmClient domain.LLMClient
}

func NewIngestHandler() IngestHandler {
	return IngestHandler{}
}

func (h IngestHandler) Handle(
	ctx context.Context,
	text string,
	documentID string,
) error {
	chunks, err := h.chunker.Chunk(text, documentID)
	if err != nil {
		return err
	}

	texts := make([]string, 0, len(chunks))
	for _, c := range chunks {
		texts = append(texts, c.Text)
	}

	embs, err := h.llmClient.Embed(ctx, texts)
	if err != nil {
		return err
	}

	if err := h.store.UpsertChunks(ctx, chunks, embs); err != nil {
		return err
	}

	return nil
}
