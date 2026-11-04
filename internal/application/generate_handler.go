package application

import (
	"context"
	"fmt"
	"strings"

	"rag-service-go/internal/domain"
)

type GenerateHandler struct {
	llmClient   domain.LLMClient
	vectorStore domain.VectorStore
}

func NewGenerateHandler(
	llmClient domain.LLMClient,
	vectorStore domain.VectorStore,
) GenerateHandler {
	return GenerateHandler{
		llmClient:   llmClient,
		vectorStore: vectorStore,
	}
}

func (h GenerateHandler) Handle(ctx context.Context, prompt string) (string, error) {
	queryEmbs, err := h.llmClient.Embed(ctx, []string{prompt})
	if err != nil {
		return "", fmt.Errorf("failed to embed query prompt: %w", err)
	}
	if len(queryEmbs) == 0 {
		return "", fmt.Errorf("no embedding returned for query prompt")
	}

	const topK = 3
	searchResults, err := h.vectorStore.Search(ctx, queryEmbs[0], topK)
	if err != nil {
		return "", fmt.Errorf("failed to search vector store: %w", err)
	}

	finalPrompt := prompt
	if len(searchResults) > 0 {
		var contextBuilder strings.Builder
		for i, res := range searchResults {
			fmt.Fprintf(&contextBuilder, "[Chunk %d]: %s\n", i+1, res.Chunk.Text)
		}

		finalPrompt = fmt.Sprintf(`Context information is below:
		---------------------
		%s---------------------
		Given the context information above and not prior knowledge, answer the following query:
		Query: %s
		Answer:`, contextBuilder.String(), prompt)
	}

	return h.llmClient.Generate(ctx, finalPrompt)
}
