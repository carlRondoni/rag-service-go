//go:build integration

package llm_clients_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"rag-service-go/internal/infrastructure/llm_clients"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getOllamaURL() string {
	if url := os.Getenv("OLLAMA_URL"); url != "" {
		return url
	}
	return "http://localhost:11434"
}

func getOllamaModel() string {
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		return model
	}
	return "llama3"
}

func getOllamaEmbedModel() string {
	if model := os.Getenv("OLLAMA_EMBED_MODEL"); model != "" {
		return model
	}
	return "nomic-embed-text"
}

func TestOllamaClient_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	baseURL := getOllamaURL()
	model := getOllamaModel()
	embedModel := getOllamaEmbedModel()
	logger := zerolog.New(os.Stdout)

	client := llm_clients.NewOllamaClient(baseURL, model, embedModel, logger)

	err := client.Health(ctx)
	require.NoError(t, err, "Ollama must be running at %s for integration tests (start via make docker-start)", baseURL)

	texts := []string{"Clean Architecture in Go"}
	embs, err := client.Embed(ctx, texts)
	require.NoError(t, err, "Failed to get embeddings from Ollama for model %s", embedModel)
	require.Len(t, embs, 1)
	assert.NotEmpty(t, embs[0].Vector, "Embedding vector should not be empty")

	response, err := client.Generate(ctx, "Answer with the single word 'PONG': PING")
	require.NoError(t, err, "Failed to generate text from Ollama")
	// with streaming on, this would only be the first fragment ("P")
	assert.Contains(t, strings.ToUpper(response), "PONG")
}
