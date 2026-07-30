package application

import (
	"context"
	"rag-service-go/internal/domain"
)

type GenerateHandler struct {
	llmClient domain.LLMClient
}

func NewGenerateHandler(
	llmClient domain.LLMClient,
) GenerateHandler {
	return GenerateHandler{
		llmClient: llmClient,
	}
}

func (h GenerateHandler) Handle(ctx context.Context, prompt string) (string, error) {
	return h.llmClient.Generate(ctx, prompt)
}
