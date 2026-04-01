package chunker

import (
	"fmt"

	"rag-service-go/internal/domain"
)

type SimpleChunker struct {
	chunkSize int
}

func NewSimpleChunker(size int) *SimpleChunker {
	return &SimpleChunker{chunkSize: size}
}

func (c *SimpleChunker) Chunk(text string, documentID string) ([]domain.Chunk, error) {

	var chunks []domain.Chunk

	runes := []rune(text)

	for i := 0; i < len(runes); i += c.chunkSize {
		end := i + c.chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunkText := string(runes[i:end])

		chunks = append(chunks, domain.Chunk{
			ID:         fmt.Sprintf("%s-%d", documentID, i),
			DocumentID: documentID,
			Text:       chunkText,
		})
	}

	return chunks, nil
}
