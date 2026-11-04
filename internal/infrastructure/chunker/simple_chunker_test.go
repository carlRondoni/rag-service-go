//go:build unit || !integration

package chunker_test

import (
	"regexp"
	"testing"

	"rag-service-go/internal/infrastructure/chunker"

	"github.com/stretchr/testify/assert"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestSimpleChunker_Chunking(t *testing.T) {
	c := chunker.NewSimpleChunker(5)

	text := "abcdefghij"
	chunks, err := c.Chunk(text, "doc-1")

	assert.NoError(t, err)
	assert.Len(t, chunks, 2)
	assert.Equal(t, "abcde", chunks[0].Text)
	assert.Equal(t, "fghij", chunks[1].Text)
	assert.Equal(t, "doc-1", chunks[0].DocumentID)
	assert.Equal(t, "doc-1", chunks[1].DocumentID)

	assert.True(t, uuidRegex.MatchString(chunks[0].ID), "Chunk ID should be a valid RFC 4122 UUID v4, got: %s", chunks[0].ID)
	assert.True(t, uuidRegex.MatchString(chunks[1].ID), "Chunk ID should be a valid RFC 4122 UUID v4, got: %s", chunks[1].ID)
	assert.NotEqual(t, chunks[0].ID, chunks[1].ID)
}

func TestSimpleChunker_EmptyText(t *testing.T) {
	c := chunker.NewSimpleChunker(100)

	chunks, err := c.Chunk("", "doc-empty")
	assert.NoError(t, err)
	assert.Empty(t, chunks)
}

func TestSimpleChunker_DefaultSize(t *testing.T) {
	c := chunker.NewSimpleChunker(0)

	text := "hello"
	chunks, err := c.Chunk(text, "doc-default")
	assert.NoError(t, err)
	assert.Len(t, chunks, 1)
	assert.Equal(t, "hello", chunks[0].Text)
}

