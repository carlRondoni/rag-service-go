package chunker

import (
	"crypto/rand"
	"fmt"

	"rag-service-go/internal/domain"
)

type SimpleChunker struct {
	chunkSize int
}

func NewSimpleChunker(size int) *SimpleChunker {
	if size <= 0 {
		size = 500
	}
	return &SimpleChunker{chunkSize: size}
}

func generateUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // RFC 4122 version 4
	b[8] = (b[8] & 0x3f) | 0x80 // RFC 4122 variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (c *SimpleChunker) Chunk(text string, documentID string) ([]domain.Chunk, error) {
	var chunks []domain.Chunk

	runes := []rune(text)
	if len(runes) == 0 {
		return chunks, nil
	}

	for i := 0; i < len(runes); i += c.chunkSize {
		end := i + c.chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunkText := string(runes[i:end])

		chunks = append(chunks, domain.Chunk{
			ID:         generateUUID(),
			DocumentID: documentID,
			Text:       chunkText,
		})
	}

	return chunks, nil
}
